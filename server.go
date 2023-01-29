package httpbase

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-nacelle/nacelle/v2"
	"github.com/go-nacelle/process/v2"
	"github.com/go-nacelle/service/v2"
	"github.com/google/uuid"
)

type (
	Server struct {
		Logger          nacelle.Logger            `service:"logger"`
		Config          *nacelle.Config           `service:"config"`
		Services        *nacelle.ServiceContainer `service:"services"`
		Health          *nacelle.Health           `service:"health"`
		tagModifiers    []nacelle.TagModifier
		initializer     ServerInitializer
		listener        *net.TCPListener
		server          *http.Server
		once            *sync.Once
		host            string
		port            int
		certFile        string
		keyFile         string
		shutdownTimeout time.Duration
		healthToken     healthToken
		healthStatus    *process.HealthComponentStatus
	}

	ServerInitializer interface {
		Init(context.Context, *http.Server) error
	}

	ServerInitializerFunc func(context.Context, *http.Server) error
)

func (f ServerInitializerFunc) Init(ctx context.Context, server *http.Server) error {
	return f(ctx, server)
}

func NewServer(initializer ServerInitializer, configs ...ConfigFunc) *Server {
	options := getOptions(configs)

	return &Server{
		tagModifiers: options.tagModifiers,
		initializer:  initializer,
		once:         &sync.Once{},
		healthToken:  healthToken(uuid.New().String()),
	}
}

func (s *Server) Init(ctx context.Context) (err error) {
	healthStatus, err := s.Health.Register(s.healthToken)
	if err != nil {
		return err
	}
	s.healthStatus = healthStatus

	httpConfig := &Config{}
	if err = s.Config.Load(httpConfig, s.tagModifiers...); err != nil {
		return err
	}

	s.listener, err = makeListener(httpConfig.HTTPHost, httpConfig.HTTPPort)
	if err != nil {
		return err
	}

	s.server = &http.Server{}
	s.host = httpConfig.HTTPHost
	s.port = httpConfig.HTTPPort
	s.certFile = httpConfig.HTTPCertFile
	s.keyFile = httpConfig.HTTPKeyFile
	s.shutdownTimeout = httpConfig.ShutdownTimeout

	if err := service.Inject(ctx, s.Services, s.initializer); err != nil {
		return err
	}

	return s.initializer.Init(ctx, s.server)
}

func (s *Server) Start() error {
	defer s.listener.Close()
	defer s.server.Close()

	s.healthStatus.Update(true)

	if s.certFile != "" {
		return s.serveTLS()
	}

	return s.serve()
}

func (s *Server) serve() error {
	s.Logger.Info("Serving HTTP on %s:%d", s.host, s.port)
	if err := s.server.Serve(s.listener); err != http.ErrServerClosed {
		return err
	}

	s.Logger.Info("No longer serving HTTP on %s:%d", s.host, s.port)
	return nil
}

func (s *Server) serveTLS() error {
	s.Logger.Info("Serving HTTP/TLS on %s:%d", s.host, s.port)
	if err := s.server.ServeTLS(s.listener, s.certFile, s.keyFile); err != http.ErrServerClosed {
		return err
	}

	s.Logger.Info("No longer serving HTTP/TLS on %s:%d", s.host, s.port)
	return nil
}

func (s *Server) Stop() (err error) {
	s.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		s.Logger.Info("Shutting down HTTP server")
		err = s.server.Shutdown(ctx)
	})

	return
}

func makeListener(host string, port int) (*net.TCPListener, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	return net.ListenTCP("tcp", addr)
}
