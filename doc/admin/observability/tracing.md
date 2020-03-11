# Tracing

## Inspecting traces (Jaeger or LightStep)

If LightStep or Jaeger is configured (using the [`useJaeger` or `lightstep*` site configuration
properties](config/site_config.md)), every HTTP response will include an `X-Trace` header with a link
to the trace for that request. Inspecting the spans and logs attached to the trace will help
identify the problematic service or dependency.


## Viewing Go net/trace information

If you are using Sourcegraph's Docker deployment, site admins can access [Go
`net/trace`](https://godoc.org/golang.org/x/net/trace) information at
https://sourcegraph.example.com/-/debug/. If you are using Sourcegraph cluster, you need to `kubectl
port-forward ${POD_NAME} 6060` to access the debug page. From there, when you are viewing the debug
page of a service, click **Requests** to view the traces for that service.
