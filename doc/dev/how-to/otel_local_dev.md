# Set up local Sourcegraph OpenTelemetry development

> WARNING: OpenTelemetry support is a work in progress, and so are these docs!

General OpenTelemetry export configuration is done via environment variables according to the [official configuration options specification](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#configuration-options).

## Collector

`sg start otel` -> runs `otel-collector` and `jaeger`. Additional configuration can be provided to export to different destinations - for example, to configure a simple Honeycomb exporter, add the following to your `sg.config.overwrite.yaml`:

```yaml
commands:
  otel-collector:
    env:
      CONFIGURATION_FILE: 'configs/honeycomb.yaml'
      HONEYCOMB_API_KEY: '...'
```

To learn more, see [`docker-images/opentelemetry-collector`](https://github.com/sourcegraph/sourcegraph/tree/main/docker-images/opentelemetry-collector).

## Configuration

Set `dev-private` site config to use `"observability.tracing": { "type": "opentelemetry" }` to enable OpenTelemetry export for most services.

To enable Zoekt OpenTelemetry, add the following to your `sg.config.overwrite.yaml`:

```yaml
commands:
  zoekt-web-0: &zoekt_otel
    JAEGER_DISABLED: true
    OPENTELEMETRY_DISABLED: false
  zoekt-web-1:
    <<: *zoekt_otel
```

## Testing

Use `sg start` to start services, and run a complex query with `&trace=1`, e.g. [`foobar(...) patterntype:structural`](https://sourcegraph.test:3443/search?q=context%3Aglobal+foobar%28...%29&patternType=structural&trace=1) - this will show the `View trace` button in the search results.

When using different backends, you can use `"urlTemplate"` in `"observability.tracing"` to configure the link.
