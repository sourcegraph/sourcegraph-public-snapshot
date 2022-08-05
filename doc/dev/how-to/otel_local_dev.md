# Set up local Sourcegraph OpenTelemetry development

> WARNING: OpenTelemetry support is a work in progress, and so are these docs!

General OpenTelemetry export configuration is done via environment variables according to the [official configuration options specification](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md#configuration-options).

## Tracing

1. Set `dev-private` site config to use `opentelemetry`
2. `sg start otel` -> runs `otel-collector` and `jaeger`
3. `sg start`
4. Run a complex query with `&trace=1`, e.g. [`foobar(...) patterntype:structural`](https://sourcegraph.test:3443/search?q=context%3Aglobal+foobar%28...%29&patternType=structural&trace=1)
5. Click `View trace`
