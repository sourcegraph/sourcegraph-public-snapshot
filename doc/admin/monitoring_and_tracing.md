# Monitoring and tracing

## Builtin monitoring

A Sourcegraph instance includes [Prometheus](https://prometheus.io/) for monitoring and [Grafana](https://grafana.com) for monitoring dashboards.
 
Site admins can view the monitoring dashboards on a Sourcegraph instance:

1. Go to **User menu > Site admin**.
1. Open the **Monitoring** page (last menu item in the left sidebar). (The URL is `https://sourcegraph.example.com/-/debug/grafana/?orgId=1`.)

See [descriptions of the Grafana dashboards provisioned by Sourcegraph](monitoring_dashboards/index.md). 

> NOTE: We are running Grafana behind a reverse proxy. Grafana is not fully integrated with our CSRF protection so there is a known issue: when the Grafana
> web app in the browser makes POST or PUT requests Sourcegraph's CSRF protection gets triggered and responds with a "invalid CSRF token" 403 response.
> We are working to solve [this issue](https://github.com/sourcegraph/sourcegraph/issues/6075). As a workaround, site admins can connect to Grafana directly to make changes to the dashboards. 
>If you're using the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph), 
>see "[Kubernetes cluster administrator guide](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/admin-guide.md)" and
> "[Grafana README](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/grafana/README.md)" for more information.

## Additional monitoring and tracing systems

Sourcegraph supports forwarding internal performance and debugging information to many monitoring and tracing systems.

- [Jaeger](https://github.com/jaegertracing/jaeger#readme) tracing, configured via the `useJaeger` properties in [critical configuration](config/critical_config.md)
- [LightStep](https://lightstep.com) tracing, configured via the `lightstep*` properties in [critical configuration](config/critical_config.md) (full [OpenTracing](http://opentracing.io/) support coming soon)
- [Sentry](https://sentry.io) logging, configured via the `sentry` property in [critical configuration](config/critical_config.md)
- [Go net/trace](#viewing-go-net-trace-information)
- [Honeycomb](https://honeycomb.io/)

If you're using the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph), see "[Kubernetes cluster administrator guide](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/admin-guide.md)" and "[Prometheus README](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/prometheus/README.md)" for more information.

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
