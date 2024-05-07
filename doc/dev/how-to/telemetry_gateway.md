# How to set up Telemetry Gateway locally

> WARNING: This is a guide intended for development reference.
>
> To learn more about telemetry export, refer to the [Sourcegraph adminstrator documentation on telemetry](../../admin/telemetry/index.md).

Telemetry Gateway is a managed service that ingests events exported from Sourcegraph instances, manipulates them as needed, and exports them to designated Pub/Sub topics or other destinations for processing.

It exposes a gRPC API defined in [`telemetrygateway/v1`](https://github.com/sourcegraph/sourcegraph/tree/main/internal/telemetrygateway/v1), and the service itself is implemented in [`cmd/telemetry-gateway`](https://github.com/sourcegraph/sourcegraph/tree/main/cmd/telemetry-gateway).

To learn more about the Sourcegraph's new Telemetry framework, refer to [the telemetry documentation](../background-information/telemetry/index.md).

> NOTE: In the Sourcegraph application, the [new events being exported using `internal/telemetry`](../background-information/telemetry/index.md) are sometimes loosely referred to as "V2", as it supersedes the existing mechanisms of writing directly to the `event_logs` database table.
> The *Telemetry Gateway* schema, however, is `telemetrygateway/v1`, as it is the first iteration of the service's API.

## Default development behaviour

A test deployment is available at `telemetry-gateway.sgdev.org` (see [go/msp-ops/telemetry-gateway#dev](https://handbook.sourcegraph.com/departments/engineering/managed-services/telemetry-gateway/#dev)), which publishes events to a test topic and development pipeline - currently [`sourcegraph-telligent-testing/event-telemetry-test`](https://console.cloud.google.com/cloudpubsub/topic/edit/event-telemetry-test?project=sourcegraph-telligent-testing).
This instance only accepts licensed instance events that use a development-only license key, and is continuously deployed using MSP rollouts.

Exports of [V2 telemetry events](../background-information/telemetry/index.md) to this development instance is enabled by default in development using the `TELEMETRY_GATEWAY_EXPORTER_EXPORT_ADDR` environment variable configured in `sg.config.yaml` - for example, `sg start` will export V2 telemetry events to this instance.

## Running Telemetry Gateway locally

First, start a Telemetry Gateway instance locally:

```sh
sg run telemetry-gateway
```

Then, configure the `TELEMETRY_GATEWAY_EXPORTER_EXPORT_ADDR` environment variable in `sg.config.overwrite.yaml` to send events to this locally running instance:

```yaml
env:
  TELEMETRY_GATEWAY_EXPORTER_EXPORT_ADDR: 'http://127.0.0.1:6080'
```

By default, the local Telemetry Gateway instance will simply log any events it receives at `debug` level without forwarding the events anywhere.
To see the message payloads it *would* emit in a production environment, configure the log level in `sg.config.overwrite.yaml` as well:

```yaml
commands:
  telemetry-gateway:
    env:
      SRC_LOG_LEVEL: debug
```

You can increase the frequency of exports to monitor behaviour closer to real-time by setting the following in `sg.config.yaml`:

```yaml
env:
  TELEMETRY_GATEWAY_EXPORTER_EXPORT_INTERVAL: "10s"
```

In development, a gRPC interface is enabled for Telemetry Gateway as well at `http://127.0.0.1:10085/debug/grpcui/`.
