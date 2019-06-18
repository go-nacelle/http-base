package main

import (
	"net/http"

	"github.com/go-nacelle/httpbase"
	"github.com/go-nacelle/nacelle"
)

type ServerInitializer struct{}

func (si *ServerInitializer) Init(config nacelle.Config, server *http.Server) error {
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!\n"))
	})

	return nil
}

func setup(processes nacelle.ProcessContainer, services nacelle.ServiceContainer) error {
	processes.RegisterProcess(httpbase.NewServer(&ServerInitializer{}))
	return nil
}

func main() {
	nacelle.NewBootstrapper("httpbase-example", setup).BootAndExit()
}
