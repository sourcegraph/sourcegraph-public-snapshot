# Tracing

## Jaeger

If Jaeger is configured (using the `useJaeger` [site configuration](../config/site_config.md))
property, every HTTP response will include an `X-Trace` header with a link to the trace for that
request. Inspecting the spans and logs attached to the trace will help identify the problematic
service or dependency.

### Set up Jaeger

If you are using the single Docker container or Docker Compose deployment method for Sourcegraph,
refer to the [official Jaeger documentation](https://www.jaegertracing.io/) for
guidance. Note that you have two options:

* Run Jaeger using the
  [all-in-one](https://www.jaegertracing.io/docs/1.16/getting-started/#all-in-one) container
* Run the constituent services of Jaeger (jaeger-agent, jaeger-collector, jaeger-query, optionally
  jaeger-ingester) separately.
  * If running separately, you must choose which type of storage (e.g., Cassandra, Elasticsearch,
    memory) is used to store Jaeger spans.

If you are using Kubernetes, you can refer to [these docs on how to run Jaeger in the Sourcegraph
cluster](https://github.com/sourcegraph/deploy-sourcegraph/tree/master/configure/jaeger).

### Accessing Jaeger

If you are using the single Docker container or Docker Compose deployment, you'll need to make the
Jaeger UI (jaeger-query) accessible to site admins.

If you are using Kubernetes and have followed the recommended docs above to run Jaeger in the
Sourcegraph cluster, you can access the the Jaeger UI with port-forwarding at http://localhost:16686/:

```
kubectl port-forward svc/jaeger-query 16686
```

## Viewing Go net/trace information

Site admins can access [Go `net/trace`](https://godoc.org/golang.org/x/net/trace) information at
https://sourcegraph.example.com/-/debug/. From there, click **Requests** to view the traces for that
service.
