# Logs

## Log levels

A Sourcegraph service's log level is configured via the environment variable `SRC_LOG_LEVEL`. The valid values (from most to least verbose) are:

* `dbug`: Debug. Output all logs. Default in cluster deployments.
* `info`: Informational.
* `warn`: Warning. Default in Docker deployments.
* `eror`: Error.
* `crit`: Critical.

Learn more about how to apply these environment variables in [docker-compose](../install/docker-compose/operations.md#set-environment-variables) and [server](../install/docker/operations.md#environment-variables) deployments. 

## Log format

A Sourcegraph service's log output format is configured via the environment variable `SRC_LOG_FORMAT`. The valid values are:

* `condensed`: Optimized for human readability.
* `json`: Machine-readable JSON format.
* `logfmt`: The [logfmt](https://github.com/kr/logfmt) format.
  * Note that `logfmt` is no longer supported with [Sourcegraph's new internal logging standards](../../dev/how-to/add_and_use_logging.md) - if you need structured logs, we recommend using `json` instead. If set to `logfmt`, log output from new loggers will be in `condensed` format.
