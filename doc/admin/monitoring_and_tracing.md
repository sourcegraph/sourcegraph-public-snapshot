# Monitoring and tracing

## Builtin monitoring

A Sourcegraph instance includes [Prometheus](https://prometheus.io/) for monitoring and [Grafana](https://grafana.com) for monitoring dashboards.
 
Site admins can view the monitoring dashboards on a Sourcegraph instance:

1. Go to **User menu > Site admin**.
1. Open the **Monitoring** page (last menu item in the left sidebar). (The URL is `https://sourcegraph.example.com/-/debug/grafana/?orgId=1`.)

See [descriptions of the Grafana dashboards provisioned by Sourcegraph](monitoring_dashboards/index.md). 

> NOTE: We are running Grafana behind a reverse proxy. Grafana is not fully integrated with our CSRF protection so there is a known issue: when the Grafana
> web app in the browser makes POST or PUT requests Sourcegraph's CSRF protection gets triggered and responds with a "invalid CSRF token" 403 response.
> We are working to solve [this issue](https://github.com/sourcegraph/sourcegraph/issues/6075). 
> As a workaround, site admins can connect to Grafana directly (as described below) to make changes to the dashboards. 

## Accessing Grafana directly

Follow the instructions below to access Grafana directly by visiting http://localhost:3370/-/debug/grafana. 
This URL will show the home dashboard and from there you can add, modify and delete your own dashboards and panels,
 as well as configure alerts.

### Kubernetes

If you're using the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph),  
you can access Grafana directly using Kubernetes port forwarding to your local machine:

```bash script
kubectl port-forward svc/grafana 3370:30070
``` 

### Single-container server deployments

For simplicity, Garafana does not require authentication, as the port binding of 3370 is restricted to connections from localhost only.

Therefore, if accessing Grafana locally, the URL will be http://localhost:3370/-/debug/grafana. If Sourcegraph is deployed to a remote server, then access via an SSH tunnel using a tool
such as [sshuttle](https://github.com/sshuttle/sshuttle) is required to establish a secure connection to Grafana.
To access the remote server using `sshuttle` from your local machine:

```bash script
sshuttle -r user@host 0/0
```

Then simply visit http://host:3370 in your browser.

### Docker images

#### Prometheus

We are running our own image of Prometheus which contains a standard Prometheus installation packaged together 
with rules files and target files for our monitoring.

A directory can be mounted at `/sg_prometheus_add_ons`. It can contains additional config files of two types:
  - rule files which must have the suffix `_rules.yml` in their filename (ie `gitserver_rules.yml`)
  - target files which must have the suffix `_targets.yml` in their filename (ie `local_targets.yml`)

[Rule files](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/) 
and [target files](https://prometheus.io/docs/guides/file-sd/) must use the latest Prometheus 2.x syntax.  

The environment variable `PROMETHEUS_ADDITIONAL_FLAGS` can be used to pass on additional flags to the prometheus executable running in the container.

#### Grafana

We are running our own image of Grafana which contains a standard Grafana installation packaged together with provisioned dashboards.

> NOTE: Our Grafana instance runs in anonymous mode with all authentication turned off. Please be careful when exposing it directly.

A directory containing dashboard json specifications can be mounted in the docker container at
`/sg_grafana_additional_dashboards` and they will be picked up automatically. Changes to files in that directory
will be detected automatically while Grafana is running.

More behavior can be controlled with
[environmental variables](https://grafana.com/docs/installation/configuration/).

## Additional monitoring and tracing systems

Sourcegraph supports forwarding internal performance and debugging information to many monitoring and tracing systems.

- [Jaeger](https://github.com/jaegertracing/jaeger#readme) tracing, configured via the `useJaeger` properties in [site configuration](config/site_config.md)
- [LightStep](https://lightstep.com) tracing, configured via the `lightstep*` properties in [site configuration](config/site_config.md) (full [OpenTracing](http://opentracing.io/) support coming soon)
- [Sentry](https://sentry.io) logging, configured via the `sentry` property in [site configuration](config/site_config.md)
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

If LightStep or Jaeger is configured (using the [`useJaeger` or `lightstep*` site configuration properties](config/site_config.md), every HTTP response will include an `X-Trace` header with a link to the trace for that request. Inspecting the spans and logs attached to the trace will help identify the problematic service or dependency.

### Viewing Go net/trace information

If you are using Sourcegraph's Docker deployment, site admins can access [Go `net/trace`](https://godoc.org/golang.org/x/net/trace) information at https://sourcegraph.example.com/-/debug/. If you are using Sourcegraph cluster, you need to `kubectl port-forward ${POD_NAME} 6060` to access the debug page. From there, when you are viewing the debug page of a service, click **Requests** to view the traces for that service.
