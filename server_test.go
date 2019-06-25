package httpbase

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/aphistic/sweet"
	"github.com/go-nacelle/nacelle"
	. "github.com/onsi/gomega"
)

type ServerSuite struct{}

var testConfig = nacelle.NewConfig(nacelle.NewTestEnvSourcer(map[string]string{
	"http_port": "0",
}))

func (s *ServerSuite) TestServeAndStop(t sweet.T) {
	server := makeHTTPServer(func(config nacelle.Config, server *http.Server) error {
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

	err := server.Init(testConfig)
	Expect(err).To(BeNil())

	go server.Start()
	defer server.Stop()

	// Hack internals to get the dynamic port (don't bind to one on host)
	url := fmt.Sprintf("http://localhost:%d/users/foo", getDynamicPort(server.listener))

	req, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())

	resp, err := http.DefaultClient.Do(req)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	Expect(err).To(BeNil())
	Expect(data).To(Equal([]byte("bar")))
}

func (s *ServerSuite) TestServeTLS(t sweet.T) {
	server := makeHTTPServer(func(config nacelle.Config, server *http.Server) error {
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

	err := server.Init(nacelle.NewConfig(nacelle.NewTestEnvSourcer(map[string]string{
		"http_port":      "0",
		"http_cert_file": "./internal/keys/server.crt",
		"http_key_file":  "./internal/keys/server.key",
	})))

	Expect(err).To(BeNil())

	go server.Start()
	defer server.Stop()

	// Hack internals to get the dynamic port (don't bind to one on host)
	url := fmt.Sprintf("https://localhost:%d/users/foo", getDynamicPort(server.listener))

	req, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	Expect(err).To(BeNil())
	Expect(data).To(Equal([]byte("bar")))
}

func (s *ServerSuite) TestBadInjection(t sweet.T) {
	server := NewServer(&badInjectionHTTPInitializer{})
	server.Services = makeBadContainer()
	server.Health = nacelle.NewHealth()

	err := server.Init(testConfig)
	Expect(err.Error()).To(ContainSubstring("ServiceA"))
}

func (s *ServerSuite) TestTagModifiers(t sweet.T) {
	server := NewServer(
		ServerInitializerFunc(func(config nacelle.Config, server *http.Server) error {
			return nil
		}),
		WithTagModifiers(nacelle.NewEnvTagPrefixer("prefix")),
	)

	server.Logger = nacelle.NewNilLogger()
	server.Services = nacelle.NewServiceContainer()
	server.Health = nacelle.NewHealth()

	err := server.Init(nacelle.NewConfig(nacelle.NewTestEnvSourcer(map[string]string{
		"prefix_http_port": "1234",
	})))

	Expect(err).To(BeNil())
	Expect(server.port).To(Equal(1234))
}

func (s *ServerSuite) TestInitError(t sweet.T) {
	server := makeHTTPServer(func(config nacelle.Config, server *http.Server) error {
		return fmt.Errorf("oops")
	})

	err := server.Init(testConfig)
	Expect(err).To(MatchError("oops"))
}

//
// Helpers

func makeHTTPServer(initializer func(nacelle.Config, *http.Server) error) *Server {
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

func (i *badInjectionHTTPInitializer) Init(nacelle.Config, *http.Server) error {
	return nil
}

func makeBadContainer() nacelle.ServiceContainer {
	container := nacelle.NewServiceContainer()
	container.Set("A", &B{})
	return container
}
