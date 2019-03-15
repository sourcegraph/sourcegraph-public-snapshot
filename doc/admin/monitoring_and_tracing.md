# Monitoring and tracing

Sourcegraph supports forwarding internal performance and debugging information to many monitoring and tracing systems.

- [LightStep](https://lightstep.com) (full [OpenTracing](http://opentracing.io/) support coming soon)
- [Jaeger](https://github.com/jaegertracing/jaeger#readme)
- [Go net/trace](https://godoc.org/golang.org/x/net/trace)
- [Honeycomb](https://honeycomb.io/)
- [Prometheus](https://prometheus.io/) and alerting systems that integrate with it

If you're using the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph), see "[Kubernetes cluster administrator guide](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/admin-guide.md)" and "[Prometheus README](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/prometheus/README.md)" for more information.

We are in the process of documenting more common monitoring and tracing deployment scenarios. For help configuring monitoring and tracing on your Sourcegraph instance, contact us at [@srcgraph](https://twitter.com/srcgraph) or <mailto:support@sourcegraph.com>, or file issues on our [public issue tracker](https://github.com/sourcegraph/issues/issues).

## Health check

An application health check status endpoint is available at the URL path `/healthz`. It returns HTTP 200 if and only if the main frontend server and databases (PostgreSQL and Redis) are available.

The [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph) ships with comprehensive health checks for each Kubernetes deployment.
