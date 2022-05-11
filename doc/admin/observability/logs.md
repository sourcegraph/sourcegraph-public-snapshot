# Logs

This document describes the log output from Sourcegraph services and how to configure it.

## Log levels

A Sourcegraph service's log level is configured via the environment variable `SRC_LOG_LEVEL`. The valid values (from most to least verbose) are:

* `dbug`: Debug. Output all logs. Default in cluster deployments.
* `info`: Informational.
* `warn`: Warning. Default in Docker deployments.
* `eror`: Error.
* `crit`: Critical.

Learn more about how to apply these environment variables in [docker-compose](../deploy/docker-compose/index.md#set-environment-variables) and [server](../deploy/docker-single-container/index.md#environment-variables) deployments. 

## Log format

A Sourcegraph service's log output format is configured via the environment variable `SRC_LOG_FORMAT`. The valid values are:

* `condensed`: Optimized for human readability.
* `json`: Machine-readable JSON format.
  * For certain services and log entries, Sourcegraph exports a [OpenTelemetry-compliant log data model](#opentelemetry).
* `logfmt`: The [logfmt](https://github.com/kr/logfmt) format.
  * Note that `logfmt` is no longer supported with [Sourcegraph's new internal logging standards](../../dev/how-to/add_logging.md) - if you need structured logs, we recommend using `json` instead. If set to `logfmt`, log output from new loggers will be in `condensed` format.

### OpenTelemetry

When [configured to export JSON logs](#log-format), Sourcegraph services that have migrated to the [new internal logging standard](../../dev/how-to/add_logging.md) that will export a JSON log format compliant with [OpenTelemetry's log data model](https://opentelemetry.io/docs/reference/specification/logs/data-model/):

```json
{
  "Timestamp": 1651000257893614000,
  "InstrumentationScope": "string",
  "SeverityText": "string (DEBUG, INFO, ...)",
  "Body": "string",
  "Attributes": { "key": "value" },
  "Resource": {
    "service.name": "string",
    "service.version": "string",
    "service.instance.id": "string",
  },
  "TraceId": "string (optional)",
  "SpanId": "string (optional)",
}
```

We also include the following non-OpenTelemetry fields:

```json
{
  "Caller": "string",
  "Function": "string",
}
```
