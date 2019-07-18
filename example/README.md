# HTTP Base Example

A trivial example application to showcase the [httpbase](https://nacelle.dev/docs/base-processes/httpbase) library.

## Overview

This example application uses Redis to provide a simple string get/set API over HTTP. The **main** function boots [nacelle](https://nacelle.dev/docs/core) with a initializer that dials Redis and a server initializer for the process provided by this library. The connection created by the former is injected into the later.

## Building and Running

If running in Docker, simply run `docker-compose up`. This will compile the example application via a multi-stage build and start a container for the API as well as a container for the Redis dependency.

If running locally, simply build with `go build` (using Go 1.12 or above) and invoke with `REDIS_ADDR=redis://{your_redis_host}:6379 ./example`.

## Usage

```bash
$ curl -i http://localhost:5000/example-key
HTTP/1.1 404 Not Found
Date: Fri, 21 Jun 2019 00:59:21 GMT
Content-Length: 0
```

```bash
$ curl -i http://localhost:5000/example-key -X POST -d 'payload'
HTTP/1.1 200 OK
Date: Fri, 21 Jun 2019 00:59:30 GMT
Content-Length: 0
```

```bash
$ curl -i http://localhost:5000/example-key
HTTP/1.1 200 OK
Date: Fri, 21 Jun 2019 00:59:32 GMT
Content-Length: 7
Content-Type: text/plain; charset=utf-8

payload
```
