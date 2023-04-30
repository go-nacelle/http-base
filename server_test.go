package httpbase

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/go-nacelle/nacelle/v2"
	"github.com/stretchr/testify/assert"
)

var testConfig = nacelle.NewConfig(nacelle.NewTestEnvSourcer(map[string]string{
	"http_port": "0",
}))

func TestServeAndStop(t *testing.T) {
	server := makeHTTPServer(func(ctx context.Context, server *http.Server) error {
		server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/users/foo" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("bar"))
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
		})

		return nil
	})
	server.Config = testConfig

	ctx := context.Background()
	err := server.Init(ctx)
	assert.Nil(t, err)

	go server.Start(ctx)
	defer server.Stop(ctx)

	// Hack internals to get the dynamic port (don't bind to one on host)
	url := fmt.Sprintf("http://localhost:%d/users/foo", getDynamicPort(server.listener))

	req, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, []byte("bar"), data)
}

func TestServeTLS(t *testing.T) {
	server := makeHTTPServer(func(ctx context.Context, server *http.Server) error {
		server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/users/foo" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("bar"))
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
		})

		return nil
	})
	server.Config = nacelle.NewConfig(nacelle.NewTestEnvSourcer(map[string]string{
		"http_port":      "0",
		"http_cert_file": "./internal/keys/server.crt",
		"http_key_file":  "./internal/keys/server.key",
	}))

	ctx := context.Background()
	err := server.Init(ctx)

	assert.Nil(t, err)

	go server.Start(ctx)
	defer server.Stop(ctx)

	// Hack internals to get the dynamic port (don't bind to one on host)
	url := fmt.Sprintf("https://localhost:%d/users/foo", getDynamicPort(server.listener))

	req, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, []byte("bar"), data)
}

func TestBadInjection(t *testing.T) {
	server := NewServer(&badInjectionHTTPInitializer{})
	server.Services = makeBadContainer()
	server.Health = nacelle.NewHealth()
	server.Config = testConfig

	ctx := context.Background()
	err := server.Init(ctx)
	assert.Contains(t, err.Error(), "ServiceA")
}

func TestTagModifiers(t *testing.T) {
	server := NewServer(
		ServerInitializerFunc(func(ctx context.Context, server *http.Server) error {
			return nil
		}),
		WithTagModifiers(nacelle.NewEnvTagPrefixer("prefix")),
	)

	server.Logger = nacelle.NewNilLogger()
	server.Services = nacelle.NewServiceContainer()
	server.Health = nacelle.NewHealth()

	server.Config = nacelle.NewConfig(nacelle.NewTestEnvSourcer(map[string]string{
		"prefix_http_port": "1234",
	}))

	ctx := context.Background()
	err := server.Init(ctx)

	assert.Nil(t, err)
	assert.Equal(t, 1234, server.port)
}

func TestInitError(t *testing.T) {
	server := makeHTTPServer(func(ctx context.Context, server *http.Server) error {
		return fmt.Errorf("oops")
	})
	server.Config = testConfig

	ctx := context.Background()
	err := server.Init(ctx)
	assert.EqualError(t, err, "oops")
}

//
// Helpers

func makeHTTPServer(initializer func(context.Context, *http.Server) error) *Server {
	server := NewServer(ServerInitializerFunc(initializer))
	server.Logger = nacelle.NewNilLogger()
	server.Services = nacelle.NewServiceContainer()
	server.Health = nacelle.NewHealth()
	return server
}

func getDynamicPort(listener net.Listener) int {
	return listener.Addr().(*net.TCPAddr).Port
}

//
// Bad Injection

type A struct{ X int }
type B struct{ X float64 }

type badInjectionHTTPInitializer struct {
	ServiceA *A `service:"A"`
}

func (i *badInjectionHTTPInitializer) Init(context.Context, *http.Server) error {
	return nil
}

func makeBadContainer() *nacelle.ServiceContainer {
	container := nacelle.NewServiceContainer()
	container.Set("A", &B{})
	return container
}
