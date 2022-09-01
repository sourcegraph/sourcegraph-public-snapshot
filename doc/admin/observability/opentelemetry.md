# OpenTelemetry <span class="badge badge-experimental">Experimental</span>

Sourcegraph is currently working on implementing [OpenTelemetry](https://opentelemetry.io/). The first [signal](https://opentelemetry.io/docs/concepts/signals/) to be integrated is [tracing](./tracing.md).

## OpenTelemetry Collector

Sourcegraph handles telemetry data through the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/). This service can be configured to ingest, process, and then export telemetry data to an observability backend of choice. This approach offers a great deal of flexibility.

The Collector is deployed with a [custom image](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/docker-images/opentelemetry-collector). This image includes the following backend exporters:

- [OTLP gRPC](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter) (core)
- [OTLP HTTP](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter) (core)
- [Logging](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/loggingexporter) (core)
- [Jaeger](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/jaegerexporter) (contrib)
- [Google Cloud](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/googlecloudexporter) (contrib)
- [Loki](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/lokiexporter) (contrib)

In case you require an additional exporter from the [contrib repository](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter), please [reach out to us](https://about.sourcegraph.com/contact). 

## Configuration

The Collector is configured with a configuration YAML file. Refer to the [documentation](https://opentelemetry.io/docs/collector/configuration/) for an in-depth explanation of the parts that compose a full collector pipeline.

### Tracing

Basic configuration for each backend type is described below. Just adding a backend to the `exporters` block does not enable it. It must also be added to the `service` block.
Refer to the next snippet for a basic but complete example, which is the default configuration out-of-the-box:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:
        
exporters:
  logging:
    loglevel: info
    sampling_initial: 5
    sampling_thereafter: 200

service:
  pipelines:
    traces:
      receivers:
        - otlp
      exporters:
        - logging # The exporter name must be added here to enable it
```

#### Configure exporting to logging

Read the [documentation](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter/README.md) for all options.

> NOTE: the deployed Collector image is bundled with a [basic configuration with log exporting](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/docker-images/opentelemetry-collector/configs/logging.yaml). If this configuration serves your need, you do not have to provide a separate config. The Collector startup command can be set to `/bin/otelcol-sourcegraph --config=/etc/otel-collector/configs/logging.yaml`. This is the default setting for our deployment methods.

```yaml
exporters:
  logging:
    loglevel: info
    sampling_initial: 5
    sampling_thereafter: 200
```

#### Connect to an OTLP gRPC backend

Read the [documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/otlpexporter/README.md) for all options.

```yaml
exporters:
  otlp:
    endpoint: otelcol2:4317
    tls:
      cert_file: file.cert
      key_file: file.key
  otlp/2:
    endpoint: otelcol2:4317
    tls:
      insecure: true
```

#### Connect to an OTLP HTTP backend

Read the [documentation](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter/README.md) for all options.

```yaml
exporters:
  otlphttp:
    endpoint: https://example.com:4318/v1/traces
```

#### Connect to Jaeger

Read the [documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/jaegerexporter/README.md) for all options.  

The example below describes how to connect to the bundled Jaeger backend, if it is enabled for your deployment. Connecting to your own Jaeger instance might require additional configuration.

> NOTE: this requires the environment variable `$JAEGER_HOST` to be set on the Collector instance (i.e. the container in Kubernetes or Docker Compose).

# 

> NOTE: the deployed Collector image is bundled with a [basic configuration with Jaeger exporting](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/docker-images/opentelemetry-collector/configs/jaeger.yaml). If this configuration serves your need, you do not have to provide a separate config. The Collector startup command can be set to `/bin/otelcol-sourcegraph --config=/etc/otel-collector/configs/jaeger.yaml`. If you enable the bundled Jaeger instance in our deployment methods, this is preconfigured for you.

```yaml
exporters:
  jaeger:
    # Default Jaeger gRPC server
    endpoint: "$JAEGER_HOST:14250"
    tls:
      insecure: true
```

#### Connect to Google Cloud Trace

Read the [documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/googlecloudexporter/README.md) for all options.    

If you run Sourcegraph on a GCP workload, all requests will be authenticated automatically. The documentation describes other authentication methods.

```yaml
exporters:
  googlecloud:
    # See docs
    project: project-name # or fetched from credentials
    retry_on_failure:
      enabled: false
```

#### Connect to Loki

Read the [documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/lokiexporter/README.md) for all options.

```yaml
loki:
  endpoint: http://loki:3100/loki/api/v1/push
  labels:
    resource:
      # Allowing 'container.name' attribute and transform it to 'container_name', which is a valid Loki label name.
      container.name: "container_name"
      # Allowing 'k8s.cluster.name' attribute and transform it to 'k8s_cluster_name', which is a valid Loki label name.
      k8s.cluster.name: "k8s_cluster_name"
    attributes:
      # Allowing 'severity' attribute and not providing a mapping, since the attribute name is a valid Loki label name.
      severity: ""
      http.status_code: "http_status_code" 
    record:
      # Adds 'traceID' as a log label, seen as 'traceid' in Loki.
      traceID: "traceid"
```
