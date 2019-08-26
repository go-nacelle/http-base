# Nacelle Base HTTP Process [![GoDoc](https://godoc.org/github.com/go-nacelle/httpbase?status.svg)](https://godoc.org/github.com/go-nacelle/httpbase) [![CircleCI](https://circleci.com/gh/go-nacelle/httpbase.svg?style=svg)](https://circleci.com/gh/go-nacelle/httpbase) [![Coverage Status](https://coveralls.io/repos/github/go-nacelle/httpbase/badge.svg?branch=master)](https://coveralls.io/github/go-nacelle/httpbase?branch=master)

Abstract HTTP server process for nacelle.

---

This library supplies an abstract HTTP server [process](https://nacelle.dev/docs/core/process) whose behavior can be configured by implementing a `ServerInitializer` interface. For a more full-featured HTTP server framework built on nacelle, see [chevron](/docs/frameworks/chevron).

You can see an additional example of an HTTP process in the [example repository](https://github.com/go-nacelle/example), specifically the [server initializer](https://github.com/go-nacelle/example/blob/843979aaa86786784a1ca3646e8d0d1f69e29c65/cmd/http-api/server_initializer.go#L23).

### Process

An HTTP process is created by supplying an initializer, described [below](https://nacelle.dev/docs/base-processes/httpbase#server-initializer), that controls its behavior.

```go
server := httpbase.NewServer(NewServerInitializer(), options...)
```

### Server Initializer

A server initializer is a struct with an `Init` method that takes a config object and an [http.Server](https://golang.org/pkg/net/http/#Server) as parameters.  This method may return an error value, which signals a fatal error to the process that runs it. This method provides an extension point to register handlers to the server instance before the process accepts connections.

The following example registers an HTTP handler function to the server that will handle all incoming requests. Each request atomically increments a request counter on the containing initializer struct and returns its new value. In more complex applications, an HTTP router, such as [gorilla/mux](https://github.com/gorilla/mux) should likely be used.

```go
type Initializer struct {
    requests uint
}

func (i *Initializer) Init(config nacelle.Config, server *http.Server) error {
    server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        value := atomic.AddUint32(&i.requests)
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(fmt.Sprintf("Hello, #%d!\n", value)))
    })

    return nil
}
```

#### Initializer Function

A simple server initializer stuct that does not need additional methods, state, or dependency instances injected via a service container can use the server initializer function wrapper instead.

```go
ServerInitializerFunc(func(config nacelle.Config, server *http.Server) error {
    server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!\n"))
    })

    return nil
})
```

### Server Process Options

The following options can be supplied to the server constructor to tune its behavior.

<dl>
  <dt>WithTagModifiers</dt>
  <dd><a href="https://godoc.org/github.com/go-nacelle/httpbase#WithTagModifiers">WithTagModifiers</a> registers the tag modifiers to be used when loading process configuration (see <a href="https://godoc.org/github.com/go-nacelle/httpbase#Configuration">below</a>). This can be used to change default hosts and ports, or prefix all target environment variables in the case where more than one HTTP server is registered per application (e.g. health server and application server, data plane and control plane server).</dt>
</dd>

### Configuration

The default process behavior can be configured by the following environment variables.

| Environment Variable  | Default | Description |
| --------------------- | ------- | ----------- |
| HTTP_HOST             | 0.0.0.0 | The host on which to accept connections. |
| HTTP_PORT             | 5000    | The port on which to accept connections. |
| HTTP_CERT_FILE        |         | The path to the TLS cert file. |
| HTTP_KEY_FILE         |         | The path to the TLS key file. |
| HTTP_SHUTDOWN_TIMEOUT | 5       | The time (in seconds) the server can spend in a graceful shutdown. |

The one of `HTTP_CERT_FILE` and `HTTP_KEY_FILE` are set, then they must both be set. Setting these will start a TLS server.
