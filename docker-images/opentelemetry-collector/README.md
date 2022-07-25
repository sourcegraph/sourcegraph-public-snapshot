# Sourcegraph OpenTelemetry collector

This distribution of the [OpenTelemetry collector](https://opentelemetry.io/docs/collector/) provides some basic collector configuration that can be used out-of-the-box in `/etc/otel-collector/configs` with the `--config` flag.
You can also mount your own configuration to provide to the `--config` flag.

In the out-of-the-box configurations, debug pages ("zPages") are available at port 55679 by default - see [Exposed zPages routes](https://github.com/open-telemetry/opentelemetry-collector/blob/main/extension/zpagesextension/README.md#exposed-zpages-routes).
The available out-of-the-box configurations are:

- [`configs/jaeger.yaml`](configs/jaeger.yaml) - useful for sending traces to Sourcegraph's bundled Jaeger instance.
- [`configs/honeycomb.yaml`](configs/honeycomb.yaml)

Each configuration requires environment variables to configure certain values - refer to the configuration file comments for more details.
To learn more about configuration in general, see the official [collector configuration docs](https://opentelemetry.io/docs/collector/configuration).
