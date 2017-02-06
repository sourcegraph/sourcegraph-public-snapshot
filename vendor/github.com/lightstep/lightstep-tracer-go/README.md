# lightstep-tracer-go

[![Circle CI](https://circleci.com/gh/lightstep/lightstep-tracer-go.svg?style=shield)](https://circleci.com/gh/lightstep/lightstep-tracer-go)
[![MIT license](http://img.shields.io/badge/license-MIT-blue.svg)](http://opensource.org/licenses/MIT)

The LightStep distributed tracing library for Go.

## Installation

```
$ go get 'github.com/lightstep/lightstep-tracer-go'
```

## Getting started

To initialize the LightStep library in particular, either retain a reference to
the LightStep `opentracing.Tracer` implementation and/or set the global
`Tracer` like so:

```
import (
    "github.com/opentracing/opentracing-go"
    "github.com/lightstep/lightstep-tracer-go"
)

func main() {
    // Initialize the LightStep Tracer; see lightstep.Options for tuning, etc.
    lightstepTracer := lightstep.NewTracer(lightstep.Options{
        AccessToken: "YourAccessToken",
    })

    // Optionally set the opentracing global Tracer to the above
    opentracing.InitGlobalTracer(lightstepTracer)

    ...
}
```

For instrumentation documentation, see the [opentracing-go
godocs](https://godoc.org/github.com/opentracing/opentracing-go).
