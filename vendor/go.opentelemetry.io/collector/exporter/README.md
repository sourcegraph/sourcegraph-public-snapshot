# General Information

An exporter defines how the pipeline data leaves the collector.

This repository hosts the following exporters available in 
traces, metrics and logs pipelines (sorted alphabetically):

- [Debug](debugexporter/README.md)
- [OTLP gRPC](otlpexporter/README.md)
- [OTLP HTTP](otlphttpexporter/README.md)

The [contrib
repository](https://github.com/open-telemetry/opentelemetry-collector-contrib)
has more exporters available in its builds.

## Configuring Exporters

Exporters are configured via YAML under the top-level `exporters` tag.

The following is a sample configuration for the `exampleexporter`.

```yaml
exporters:
  # Exporter 1.
  # <exporter type>:
  exampleexporter:
    # <setting one>: <value one>
    endpoint: 1.2.3.4:8080
    # ...
  # Exporter 2.
  # <exporter type>/<name>:
  exampleexporter/settings:
    # <setting two>: <value two>
    endpoint: 0.0.0.0:9211
```

An exporter instance is referenced by its full name in other parts of the config,
such as in pipelines. A full name consists of the exporter type, '/' and the
name appended to the exporter type in the configuration. All exporter full names
must be unique.

For the example above:

- Exporter 1 has full name `exampleexporter`.
- Exporter 2 has full name `exampleexporter/settings`.

Exporters are enabled upon being added to a pipeline. For example:

```yaml
service:
  pipelines:
    # Valid pipelines are: traces, metrics or logs
    # Trace pipeline 1.
    traces:
      receivers: [examplereceiver]
      processors: []
      exporters: [exampleexporter, exampleexporter/settings]
    # Trace pipeline 2.
    traces/another:
      receivers: [examplereceiver]
      processors: []
      exporters: [exampleexporter, exampleexporter/settings]
```

## Data Ownership

When multiple exporters are configured to send the same data (e.g. by configuring multiple
exporters for the same pipeline):
* exporters *not* configured to mutate the data will have shared access to the data
* exporters with the Capabilities to mutate the data will receive a copy of the data

Exporters access export data when `ConsumeTraces`/`ConsumeMetrics`/`ConsumeLogs`
function is called. Unless exporter's capabalities include mutation, the exporter MUST NOT modify the `pdata.Traces`/`pdata.Metrics`/`pdata.Logs` argument of
these functions. Any approach that does not mutate the original `pdata.Traces`/`pdata.Metrics`/`pdata.Logs` is allowed without the mutation capability.

## Proxy Support

Beyond standard YAML configuration as outlined in the individual READMEs above,
exporters that leverage the net/http package (all do today) also respect the
following proxy environment variables:

- HTTP_PROXY
- HTTPS_PROXY
- NO_PROXY

If set at Collector start time then exporters, regardless of protocol,
will or will not proxy traffic as defined by these environment variables.
