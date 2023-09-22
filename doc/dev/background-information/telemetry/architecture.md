# Telemetry export architecture

> WARNING: This is a guide intended for development reference.
>
> Additionally, export capabilities are **not yet enabled by default**.

This page outlines the architecture and components involved in Sourcegraph's new telemetry export system.

## Storing events

Once [recorded](./index.md#recording-events), telemetry events are stored in two places:

1. The structured `event_logs` table, for use in [admin analytics](../../../admin/analytics.md).
2. The unstructured `telemetry_events_export_queue` table, which stores raw event payloads in Protobuf wire format for export.

## Exporting events

The [`telemetrygatewayexporter`](https://github.com/sourcegraph/sourcegraph/blob/main/enterprise/cmd/worker/internal/telemetrygatewayexporter/telemetrygatewayexporter.go) running in the worker service spawns a set of background jobs that handle:

1. Reporting metrics on the `telemetry_events_export_queue`
2. Cleaning up already-exported entries in the `telemetry_events_export_queue`
3. Exporting batches of not-yet-exported entries in the `telemetry_events_export_queue` to the Telemetry Gateway service

When exporting events, we explicitly only mark an event as successfully exported when the Telemetry Gateway returns a response with a particular event's generated ID. This ensures we always export events at least once.

Note that before export, [sensitive attributes are stripped](./index.md#sensitive-attributes).

## Telemetry Gateway

The Telemetry Gateway is a managed Sourcegraph service that ingests event exports from all Sourcegraph instances, and handles manipulating the events and publishing raw payloads to a Pub/Sub topic.
It exposes a gRPC API defined in [`telemetrygateway/v1`](https://github.com/sourcegraph/sourcegraph/tree/main/internal/telemetrygateway/v1).

Also see [How to set up Telemetry Gateway locally](../../how-to/telemetry_gateway.md).
