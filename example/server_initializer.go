package main

import (
	"io/ioutil"
	"net/http"

	"github.com/garyburd/redigo/redis"
	"github.com/go-nacelle/httpbase"
	"github.com/go-nacelle/nacelle"
)

type ServerInitializer struct {
	Logger nacelle.Logger `service:"logger"`
	Redis  redis.Conn     `service:"redis"`
}

func NewServerInitializer() httpbase.ServerInitializer {
	return &ServerInitializer{}
}

func (si *ServerInitializer) Init(config nacelle.Config, server *http.Server) error {
	server.Handler = http.HandlerFunc(si.handle)
	return nil
}

func (si *ServerInitializer) handle(w http.ResponseWriter, r *http.Request) {
	if id := r.URL.Path[1:]; id != "" {
		if r.Method == "GET" {
			si.handleGet(w, r, id)
			return
		}

		if r.Method == "POST" {
			si.handlePost(w, r, id)
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (si *ServerInitializer) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	reply, err := redis.String(si.Redis.Do("GET", id))
	if err != nil {
		if err == redis.ErrNil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		si.error(w, "Failed to perform GET (%s)", err)
		return
	}

	si.Logger.Debug("Retrieved key %s", id)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(reply))
}

func (si *ServerInitializer) handlePost(w http.ResponseWriter, r *http.Request, id string) {
	defer r.Body.Close()
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		si.error(w, "Failed to read request body (%s)", err)
		return
	}

	if _, err := si.Redis.Do("SET", id, string(content)); err != nil {
		si.error(w, "Failed to perform SET (%s)", err)
		return
	}

	si.Logger.Debug("Set key %s", id)
	w.WriteHeader(http.StatusOK)
}

func (si *ServerInitializer) error(w http.ResponseWriter, format string, err error) {
	si.Logger.Error(format, err.Error())
	w.WriteHeader(http.StatusInternalServerError)
}
