# lightstep-tracer-go

## Status

The LightStep OpenTracing library makes it easy to track and visualize trace
data from OpenTracing-instrumented systems. The LightStep bindings are v0.9: we
are unaware of significant stability bugs but there are still improvements to
be made, especially regarding the OpenTracing Inject/Join functionality.

## OpenTracing API Documentation

For instrumentation documentation, see the [opentracing-go
godocs](https://godoc.org/github.com/opentracing/opentracing-go).

## Installation

```
$ go get 'github.com/lightstep/lightstep-tracer-go'
```

## Binding OpenTracing to LightStep's `Tracer`

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
