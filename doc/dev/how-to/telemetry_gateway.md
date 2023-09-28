# How to set up Telemetry Gateway locally

> WARNING: This is a guide intended for development reference.

Telemetry Gateway is a managed service that ingests events exported from Sourcegraph instances, manipulates them as needed, and exports them to designated Pub/Sub topics or other destinations for processing.

It exposes a gRPC API defined in [`telemetrygateway/v1`](https://github.com/sourcegraph/sourcegraph/tree/main/internal/telemetrygateway/v1), and the service itself is implemented in [`cmd/telemetry-gateway`](https://github.com/sourcegraph/sourcegraph/tree/main/cmd/telemetry-gateway).

To learn more about the Sourcegraph's new Telemetry framework, refer to [the telemetry documentation](../background-information/telemetry/index.md).

> NOTE: In the Sourcegraph application, the [new events being exported using `internal/telemetry`](../background-information/telemetry/index.md) are sometimes loosely referred to as "V2", as it supersedes the existing mechanisms of writing directly to the `event_logs` database table.
> The *Telemetry Gateway* schema, however, is `telemetrygateway/v1`, as it is the first iteration of the service's API.

## Running Telemetry Gateway locally

Exports of [telemetry events](../background-information/telemetry/index.md) to a local Telemetry Gateway instance is enabled in as part of `sg start` and `sg start dotcom`.
By default, the local Telemetry Gateway instance will simply log any events it receives.

You can increase the frequency of exports by setting the following in `sg.config.yaml`:

```yaml
env:
  TELEMETRY_GATEWAY_EXPORTER_EXPORT_INTERVAL: "10s"
  TELEMETRY_GATEWAY_EXPORTER_EXPORTED_EVENTS_RETENTION: "5m"
```

In development, a gRPC interface is enabled for Telemetry Gateway as well at `http://127.0.0.1:10085/debug/grpcui/`.

## Testing against a remote Telemetry Gateway

A test deployment is available at `telemetry-gateway.sgdev.org`, which publishes events to a test dataset.
In local development, you can configure Sourcegraph to export to this test deployment by setting the following in `sg.config.yaml`:

```yaml
env:
  TELEMETRY_GATEWAY_EXPORTER_EXPORT_ADDR: "https://telemetry-gateway.sgdev.org:443"
```
