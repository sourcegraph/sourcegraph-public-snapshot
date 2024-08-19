connect-opentelemetry-go
========================

[![Build](https://github.com/bufbuild/connect-opentelemetry-go/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/bufbuild/connect-opentelemetry-go/actions/workflows/ci.yaml)
[![Report Card](https://goreportcard.com/badge/github.com/bufbuild/connect-opentelemetry-go)](https://goreportcard.com/report/github.com/bufbuild/connect-opentelemetry-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/bufbuild/connect-opentelemetry-go.svg)][godoc]

`connect-opentelemetry-go` adds support for [OpenTelemetry][opentelemetry.io]
tracing and metrics collection to [connect-go] servers and clients.

For more on Connect, OpenTelemetry, and `otelconnect`, see the [Connect
announcement blog post][blog] and the observability documentation on
[connect.build](https://connect.build/docs/go/observability/).

## An example

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	connect "github.com/bufbuild/connect-go"
	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	// Generated from your protobuf schema by protoc-gen-go and
	// protoc-gen-connect-go.
	pingv1 "github.com/bufbuild/connect-opentelemetry-go/internal/gen/observability/ping/v1"
	"github.com/bufbuild/connect-opentelemetry-go/internal/gen/observability/ping/v1/pingv1connect"
)

func main() {
	mux := http.NewServeMux()

	// otelconnect.NewInterceptor provides an interceptor that adds tracing and
	// metrics to both clients and handlers. By default, it uses OpenTelemetry's
	// global TracerProvider and MeterProvider, which you can configure by
	// following the OpenTelemetry documentation. If you'd prefer to avoid
	// globals, use otelconnect.WithTracerProvider and
	// otelconnect.WithMeterProvider.
	mux.Handle(pingv1connect.NewPingServiceHandler(
		&pingv1connect.UnimplementedPingServiceHandler{},
		connect.WithInterceptors(otelconnect.NewInterceptor()),
	))

	http.ListenAndServe("localhost:8080", mux)
}

func makeRequest() {
	client := pingv1connect.NewPingServiceClient(
		http.DefaultClient,
		"http://localhost:8080",
		connect.WithInterceptors(otelconnect.NewInterceptor()),
	)
	resp, err := client.Ping(
		context.Background(),
		connect.NewRequest(&pingv1.PingRequest{}),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
}

```

## Configuration for internal services

By default, instrumented servers are conservative and behave as though they're
internet-facing. They don't trust any tracing information sent by the client,
and will create new trace spans for each request. The new spans are linked to
the remote span for reference (using OpenTelemetry's
[`trace.Link`](https://pkg.go.dev/go.opentelemetry.io/otel/trace#Link)), but
tracing UIs will display the request as a new top-level transaction.

If your server is deployed as an internal service, configure `otelconnect` to
trust the client's tracing information using
[`otelconnect.WithTrustRemote`][WithTrustRemote]. With this option, servers
will create child spans for each request.

## Reducing metrics and tracing cardinality

By default, the [OpenTelemetry RPC conventions][otel-rpc-conventions] produce
high-cardinality server-side metric and tracing output. In particular, servers
tag all metrics and trace data with the server's IP address and the remote port
number. To drop these attributes, use
[`otelconnect.WithoutServerPeerAttributes`][WithoutServerPeerAttributes]. For
more customizable attribute filtering, use
[otelconnect.WithFilter][WithFilter].

## Status

|         | Unary | Streaming Client | Streaming Handler |
|---------|:-----:|:----------------:|:-----------------:|
| Metrics | ✅    | ✅               | ✅                |
| Tracing | ✅    | ✅               | ✅                |

## Ecosystem

* [connect-go]: Service handlers and clients for GoLang
* [connect-swift]: Swift clients for idiomatic gRPC & Connect RPC
* [connect-kotlin]: Kotlin clients for idiomatic gRPC & Connect RPC
* [connect-web]: TypeScript clients for web browsers
* [Buf Studio]: web UI for ad-hoc RPCs
* [connect-crosstest]: gRPC and gRPC-Web interoperability tests

## Support and Versioning

`connect-opentelemetry-go` supports:

* The [two most recent major releases][go-support-policy] of Go.
* v1 of the `go.opentelemetry.io/otel` tracing and metrics SDK.

## Legal

Offered under the [Apache 2 license][license].

[Buf Studio]: https://buf.build/studio
[Getting Started]: https://connect.build/docs/go/getting-started
[WithFilter]: https://pkg.go.dev/github.com/bufbuild/connect-opentelemetry-go#WithFilter
[WithTrustRemote]: https://pkg.go.dev/github.com/bufbuild/connect-opentelemetry-go#WithTrustRemote
[WithoutServerPeerAttributes]: https://pkg.go.dev/github.com/bufbuild/connect-opentelemetry-go#WithoutServerPeerAttributes
[blog]: https://buf.build/blog/connect-a-better-grpc
[connect-crosstest]: https://github.com/bufbuild/connect-crosstest
[connect-go]: https://github.com/bufbuild/connect-go
[connect-kotlin]: https://github.com/bufbuild/connect-kotlin
[connect-swift]: https://github.com/bufbuild/connect-swift
[connect-web]: https://www.npmjs.com/package/@bufbuild/connect-web
[demo]: https://github.com/bufbuild/connect-demo
[docs]: https://connect.build
[go-support-policy]: https://golang.org/doc/devel/release#policy
[godoc]: https://pkg.go.dev/github.com/bufbuild/connect-opentelemetry-go
[license]: https://github.com/bufbuild/connect-opentelemetry-go/blob/main/LICENSE
[opentelemetry.io]: https://opentelemetry.io/
[otel-go-quickstart]: https://opentelemetry.io/docs/instrumentation/go/getting-started/
[otel-go]: https://github.com/open-telemetry/opentelemetry-go
[otel-rpc-conventions]: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/rpc.md
