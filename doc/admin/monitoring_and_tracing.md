# Monitoring and tracing

Sourcegraph supports forwarding internal performance and debugging information to many monitoring and tracing systems.

- [Jaeger](https://github.com/jaegertracing/jaeger#readme) tracing, configured via the `useJaeger` properties in [critical configuration](config/critical_config.md)
- [LightStep](https://lightstep.com) tracing, configured via the `lightstep*` properties in [critical configuration](config/critical_config.md) (full [OpenTracing](http://opentracing.io/) support coming soon)
- [Sentry](https://sentry.io) logging, configured via the `sentry` property in [critical configuration](config/critical_config.md)
- [Go net/trace](#viewing-go-net-trace-information)
- [Honeycomb](https://honeycomb.io/)
- [Prometheus](https://prometheus.io/) and alerting systems that integrate with it

If you're using the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph), see "[Kubernetes cluster administrator guide](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/admin-guide.md)" and "[Prometheus metrics](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/prom-metrics.md)" for more information.

We are in the process of documenting more common monitoring and tracing deployment scenarios. For help configuring monitoring and tracing on your Sourcegraph instance, use our [public issue tracker](https://github.com/sourcegraph/issues/issues).

## Health check

An application health check status endpoint is available at the URL path `/healthz`. It returns HTTP 200 if and only if the main frontend server and databases (PostgreSQL and Redis) are available.

The [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph) ships with comprehensive health checks for each Kubernetes deployment.

## Troubleshooting

Sourcegraph provides tracing, metrics and logs to help you troubleshoot problems. When investigating an issue, we recommend using the following resources:

1. [View verbose logs](#viewing-logs) (most common)
1. [Inspect traces](#inspecting-traces-jaeger-or-lightstep)
1. [Inspect the Go net/trace information](#viewing-go-net-trace-information) for individual services (rarely needed)

### Viewing logs

A Sourcegraph service's log level is configured via the environment variable `SRC_LOG_LEVEL`. The valid values (from most to least verbose) are:

* `dbug`: Debug. Output all logs. Default in cluster deployments.
* `info`: Informational.
* `warn`: Warning. Default in Docker deployments.
* `eror`: Error.
* `crit`: Critical.

If you are having issues with repository syncing, view the output of `repo-updater`'s logs.

### Inspecting traces (Jaeger or LightStep)

If LightStep or Jaeger is configured (using the [`useJaeger` or `lightstep*` critical configuration properties](config/critical_config.md), every HTTP response will include an `X-Trace` header with a link to the trace for that request. Inspecting the spans and logs attached to the trace will help identify the problematic service or dependency.

### Viewing Go net/trace information

If you are using Sourcegraph's Docker deployment, site admins can access [Go `net/trace`](https://godoc.org/golang.org/x/net/trace) information at https://sourcegraph.example.com/-/debug/. If you are using Sourcegraph cluster, you need to `kubectl port-forward ${POD_NAME} 6060` to access the debug page. From there, when you are viewing the debug page of a service, click **Requests** to view the traces for that service.
