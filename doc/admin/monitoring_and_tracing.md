# Monitoring and tracing

Sourcegraph supports forwarding internal performance and debugging information to many monitoring and tracing systems.

- [LightStep](https://lightstep.com) (full [OpenTracing](http://opentracing.io/) support coming soon)
- [Jaeger](https://github.com/jaegertracing/jaeger#readme)
- [Go net/trace](https://godoc.org/golang.org/x/net/trace)
- [Honeycomb](https://honeycomb.io/)
- [Prometheus](https://prometheus.io/) and alerting systems that integrate with it

If you're using the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph), see "[Kubernetes cluster administrator guide](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/admin-guide.md)" and "[Prometheus metrics](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/prom-metrics.md)" for more information.

We are in the process of documenting more common monitoring and tracing deployment scenarios. For help configuring monitoring and tracing on your Sourcegraph instance, contact us at [@srcgraph](https://twitter.com/srcgraph) or <mailto:support@sourcegraph.com>, or file issues on our [public issue tracker](https://github.com/sourcegraph/issues/issues).

## Health check

An application health check status endpoint is available at the URL path `/healthz`. It returns HTTP 200 if and only if the main frontend server and databases (PostgreSQL and Redis) are available.

The [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph) ships with comprehensive health checks for each Kubernetes deployment.

## Troubleshooting

Sourcegraph provides tracing, metrics and logs to facilitate troubleshooting. When investigating an issue, first [inspect traces](#tracing) (if you have tracing set up). Next, [increase the log verbosity](#logs) by setting the environment variable `SRC_LOG_LEVEL=dbug`. If that is too noisy, inspecting the Go net/trace page for individual services is valuable.

### Tracing

If Jaeger or LightStep is configured, every HTTP response will include an `X-Trace` header which links to the trace for that request. Inspecting the spans and logs attached to the trace will help quickly identify the problematic service or dependency.

### Logs

A Sourcegraph service log level is configured via the environment variable `SRC_LOG_LEVEL`. The valid values (from most to least verbose) are:

* `dbug`: Debug. Output all logs. Default in cluster deployments.
* `info`: Informational.
* `warn`: Warning. Default in Docker deployments.
* `eror`: Error.
* `crit`: Critical.

If you are having issues with repository syncing, view the output of `repo-updater`'s logs.

### Go net/trace

If you are using Sourcegraph's Docker deployment, site admins can access `net/trace` information at https://sourcegraph.example.com/-/debug/. If you are using Sourcegraph cluster, you need to `kubectl port-forward ${POD_NAME} 6060` to access the debug page. Once on the debug page of a service, click **Requests** to view the traces for that service.
