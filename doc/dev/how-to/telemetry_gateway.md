# How to set up Telemetry Gateway locally

By default, exports of Telemetry V2 events to a local Telemetry Gateway instance is enabled in `sg start` and `sg start dotcom`.

You can increase the frequency of exports by setting the following in `sg.config.yaml`:

```yaml
env:
  TELEMETRY_GATEWAY_EXPORTER_EXPORT_INTERVAL: "10s"
  TELEMETRY_GATEWAY_EXPORTER_EXPORTED_EVENTS_RETENTION: "5m"
```

In development, a gRPC interface is enabled for Telemetry Gateway as well at `http://127.0.0.1:10085/debug/grpcui/`.
