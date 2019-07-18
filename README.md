# Nacelle Base HTTP Process [![GoDoc](https://godoc.org/github.com/go-nacelle/httpbase?status.svg)](https://godoc.org/github.com/go-nacelle/httpbase) [![CircleCI](https://circleci.com/gh/go-nacelle/httpbase.svg?style=svg)](https://circleci.com/gh/go-nacelle/httpbase) [![Coverage Status](https://coveralls.io/repos/github/go-nacelle/httpbase/badge.svg?branch=master)](https://coveralls.io/github/go-nacelle/httpbase?branch=master)

Abstract HTTP server process for nacelle.

---

For a more full-featured HTTP server framework built on nacelle, see [chevron](/docs/frameworks/chevron).

### Usage

The supplied server process is an abstract HTTP/HTTPS server whose behavior is determined by a supplied `ServerInitializer` interface. This interface has only an `Init` method that receives application config as well as the HTTP server instance, allowing handlers to be registered before the server accepts clients. There is an [example](./example) included in this repository.

The following options can be supplied to the server constructor to tune its behavior.

- **WithTagModifiers** registers the tag modifiers to be used when loading process configuration (see [below](#Configuration)). This can be used to change default hosts and ports, or prefix all target environment variables in the case where more than one HTTP server is registered per application (e.g. health server and application server, data plane and control plane server).

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
