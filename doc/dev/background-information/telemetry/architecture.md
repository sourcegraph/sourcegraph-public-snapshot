# Telemetry export architecture

> WARNING: This is a guide intended for development reference.
>
> Additionally, export capabilities are **not yet enabled by default**.

This page outlines the architecture and components involved in Sourcegraph's new telemetry export system.

In the [lifecycle of an event](./index.md#event-lifecycle), events are first [stored](#storing-events) then [exported](#exporting-events) to [Telemetry Gateway](#telemetry-gateway).

See [testing events](#testing-events) for a summary of how to observe your events during development.

## Storing events

Once [recorded](./index.md#recording-events), telemetry events are stored in two places:

1. The structured `event_logs` table, for use in [admin analytics](../../../admin/analytics.md), translated from the [Telemetry Gateway format](./index.md#exported-event-schema) on a best-effort basis.
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
It exposes a gRPC API defined in [`telemetrygateway/v1`](https://github.com/sourcegraph/sourcegraph/tree/main/internal/telemetrygateway/v1) - see [exported events schema](./index.md#exported-event-schema)

Also see [How to set up Telemetry Gateway locally](../../how-to/telemetry_gateway.md).

## Testing events

In summary, when testing your events:

1. You can [see your events stored directly in `event_logs`](#storing-events) after recording.
