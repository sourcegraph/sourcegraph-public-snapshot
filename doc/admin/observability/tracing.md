# Tracing

In site configuration, you can enable tracing globally by configuring a sampling mode in `observability.tracing`.
There are currently three modes:

* `"sampling": "selective"` (default) will cause a trace to be recorded only when `trace=1` is present as a URL parameter.
* `"sampling": "all"` will cause a trace to be recorded on every request.
* `"sampling": "none"` will disable all tracing.

`"selective"` is the recommended default, because collecting traces on all requests can be quite memory- and network-intensive.
If you have a large Sourcegraph instance (e.g,. more than 10k repositories), turn this on with caution.
Note that the policies above are implemented at an application level - to sample all traces, please configure your tracing backend directly.

We support the following tracing backend types:

* [`"type": "opentelemetry"`](#opentelemetry) (default)
* [`"type": "jaeger"`](#jaeger)

In addition, we also export some tracing [via net/trace](#nettrace).

## Trace a search query

To trace a search query, run a search on your Sourcegraph instance with the `?trace=1` query parameter.
A link to the [exported trace](#tracing-backends) should be show up in the search results:

![link to trace](https://user-images.githubusercontent.com/23356519/184953302-099bcb62-ccdb-4eed-be5d-801b7fe16d97.png)

## Trace GraphQL requests

To receive a traceID on a GraphQL request, include the header `X-Sourcegraph-Should-Trace: true` with the request.
The response headers of the response will now include an `x-trace` entry, which will have a URL the [exported trace](#tracing-backends).

## Tracing backends

Tracing backends can be configured for Sourcegraph to export traces to.
We support exporting traces via [OpenTelemetry](#opentelemetry) (recommended), or directly to [Jaeger](#jaeger).

### OpenTelemetry

To learn about exporting traces to various backends using OpenTelemetry, review our [OpenTelemetry documentation](./opentelemetry.md).
Once configured, you can set up a `urlTemplate` that points to your traces backend.
For example, if you [export your traces to Honeycomb](./opentelemetry.md#otlp-compatible-backends), your configuration might look like:

```json
{
  "observability.tracing": {
    "type": "opentelemetry",
    "urlTemplate": "https://ui.honeycomb.io/$ORG/environments/$DATASET/trace?trace_id={{ .TraceID }}"
  }
}
```

You can test the exporter by [tracing a search query](#trace-a-search-query).

### Jaeger

To configure Jaeger, first ensure Jeager is running:

* **Single Docker container:** Deploy a separate Jaeger instance and configure it with [Jaeger client environment variables](https://github.com/jaegertracing/jaeger-client-go#environment-variables).
* **Docker Compose:** See the relevant [enable the bundled Jaeger deployment guide](../deploy/docker-compose/operations.md#enable-the-bundled-jaeger-deployment)
* **Kubernetes:** See the relevant [enable the bundled Jaeger deployment guide](../deploy/kubernetes/operations.md#enable-the-bundled-jaeger-deployment)

The Jaeger UI should look something like this:

![Jaeger UI](https://user-images.githubusercontent.com/1646931/79700938-0586c600-824e-11ea-9c8c-a115df8b3a21.png)

Then, configure Jaeger as your tracing backend in site configuration:

```json
{
  "observability.tracing": {
    "type": "jaeger",
    "urlTemplate": "{{ .ExternalURL }}/-/debug/jaeger/trace/{{ .TraceID }}"
  }
}
```

You can test the exporter by [tracing a search query](#trace-a-search-query).

#### Jaeger debugging algorithm

Jaeger is a powerful debugging tool that can break down where time is spent over the lifecycle of a
request and help pinpoint the source of high latency or errors. We generally follow the following
algorithm to root-cause issues with Jaeger:

1. Reproduce a slower user request (e.g., a search query that takes too long or times out).
1. Add `?trace=1` to the slow URL and reload the page, so that traces will be collected.
1. Open Chrome developer tools to the Network tab and find the corresponding GraphQL request that
   takes a long time. If there are multiple requests that take a long time, investigate them one by
   one.
1. In the Response Headers for the slow GraphQL request, find the `x-trace` header. It should
   contain a trace ID like `7edb43f744c42fbf`.
1. Go to the Jaeger UI and paste in the trace ID to the "Lookup by Trace ID" input in the top menu
   bar.
1. Explore the breakdown of the request tree in the Jaeger UI. Look for items near the leaves that
   take up a significant portion of the overall request time.
1. Report this information to Sourcegraph by screenshotting the relevant trace or by downloading the
   trace JSON.

### net/trace

Sourcegraph uses the [`net/trace`](https://pkg.go.dev/golang.org/x/net/trace) package in its backend
services, in addition to the other tracing mechanisms listed above.
This provides simple tracing information within a single process.
It can be used as an alternative when Jaeger is not available or as a supplement to Jaeger.

Site admins can access `net/trace` information at https://sourcegraph.example.com/-/debug/. From
there, click **Requests** to view the traces for that service.
