# How to add and use logging

> NOTE: Sourcegraph's new logging standards are still a work in progress - please leave a comment in [this discussion](https://github.com/sourcegraph/sourcegraph/discussions/33248) if you have any feedback or ideas!

The recommended logger at Sourcegraph is `github.com/sourcegraph/sourcegraph/lib/log`, which exports:

1. A standardized, strongly-typed, and structured logging interface, `log.Logger`
   1. Production output from this logger (`SRC_LOG_FORMAT=json`) complies with the [OpenTelemetry log data model](https://opentelemetry.io/docs/reference/specification/logs/data-model/)
   2. `log.Logger` has a variety of constructors for spawning new loggers with context, namely `Scoped`, `WithTrace`, and `WithFields`.
2. An initialization function to be called in `func main()`, `log.Init`
   1. Log level can be configured with `SRC_LOG_LEVEL`
   2. Do not use this in an `init()` function
3. A getter to retrieve a `log.Logger` instance, `log.Scoped`
   1. In development, `log.Scoped` will panic if `log.Init` is not called. In production, a no-op logger will be returned.

For testing purposes, we also provide:

1. An optional initialization function to be called in `func TestMain(*testing.M)`, `logtest.Init`
2. A getter to retrieve a `log.Logger` instance and a callback to programmatically iterate log output, `logtest.Scoped`
   1. The standard `logtest.Scoped` will also work after `logtest.Init`
   2. Programatically iterable logs are available from the `logtest.Captured` logger instance

## Core concepts

1. `lib/log` intentionally does not export directly usable log functions. Users should hold their own references to loggers, so that fields can be attached and log output can be more useful with additional context, and pass `Logger` instances to components as required.
   1. Do not call `log.Scoped` wherever you need to log - consider passing a `Logger` around explicitly.
   2. [Do not add `Logger` to `context.Context`](https://dave.cheney.net/2017/01/26/context-is-for-cancelation).
   3. [Do not create package-level logger instances](https://dave.cheney.net/2017/01/23/the-package-level-logger-anti-pattern).
2. `lib/log` should export everything that is required for logging - do not directly import a third-party logging package such as `zap`, `log15`, or `log`.

## Handling loggers

Initialize `lib/log` package within your program's `main()` function, for example:

```go
import (
  "github.com/sourcegraph/sourcegraph/lib/log"
  "github.com/sourcegraph/sourcegraph/internal/env"
  "github.com/sourcegraph/sourcegraph/internal/version"
  "github.com/sourcegraph/sourcegraph/internal/hostname"
)

func main() {
  // If unintialized, calls to `log.Scoped` will return a no-op logger in production, or
  // panic in development. It returns a callback to flush the logger buffer, if any, that
  // you should make sure to call before application exit (namely via `defer`)
  //
  // Repeated calls to `log.Init` will panic. Make sure to call this exactly once in `main`!
  syncLogs := sglog.Init(sglog.Resource{
    Name:       env.MyName,
    Version:    version.Version(),
    InstanceID: hostname.Get(),
  })
  defer syncLogs()

  service.Start(/* ... */)
}
```

When your service starts logging, get a `Logger` instance, attach some relevant context, and start propagating your logger for use:

```go
import "github.com/sourcegraph/sourcegraph/lib/log"

func newWorker(/* ... */) *Worker {
  logger := log.Scoped("worker", "the worker process handles ...").
    With(log.String("name", options.Name))
  // ...
  return &Worker{
    logger: logger,

    /* ... */
  }
}

func (w *Worker) DoSomething(params ...int) {
  _, err := doTheThing()
  if err != nil {
    // Use the provided logger instance
    w.logger.Warn("Failed to do the thing",
      log.Ints("params", params),
      log.Error(err))
  }
}
```

If you are kicking off a long-running process, you can spawn a child logger and use it directly:

```go
func (w *Worker) DoBigThing(params ...int) {
  doLog := w.logger.WithTrace(/* ... */).With("params", params)

  // subsequent entries will have trace and params attached
  doLog.Info("starting the big thing")

  // pass the logger to maintain context across function boundaries
  doSubTask(doLog, /*... */)
}
```

## Development usage

With `SRC_DEVELOPMENT=true` and `SRC_LOG_FORMAT=condensed` or `SRC_LOG_FORMAT=console`, loggers will generate a human-readable summary format like the following:

```none
DEBUG   TestInitLogger  log/logger_test.go:15   a debug message {"Attributes": {}}
INFO    TestInitLogger  log/logger_test.go:18   hello world     {"Attributes": {"some": "field", "hello": "world"}}
INFO    TestInitLogger  log/logger_test.go:21   goodbye {"TraceId": "asdf", "Attributes": {"some": "field", "world": "hello"}}
WARN    TestInitLogger  log/logger_test.go:22   another message {"TraceId": "asdf", "Attributes": {"some": "field"}}
```

This format omits fields like OpenTelemetry Resource and renders certain field types in a more friendly manner. Levels are also coloured, and the caller link with `filename:line` should be clickable in iTerm and VS Code such that you can jump straight to the source of the log entry.

## Testing usage

In the absense of `log.Init` in `main()`, `lib/log` can be initialized using `libtest` in packages that use `log.Scoped`:

```go
import (
  "os"
  "testing"

  "github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestMain(m *testing.M) {
  logtest.Init(m)
  os.Exit(m.Run())
}
```

You can control the log level used during initialization with `logtest.InitWithLevel`.

If the code you are testing accepts `Logger` instances as a parameter, you can skip the above and simply use `logtest.Scoped` to instantiate a `Logger` to provide. You can also use `logtest.Captured`, which also provides a callback for exporting logs, though be wary of making overly strict assertions on log entries to avoid brittle tests:

```go
import (
  "testing"

  "github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestFooBar(t *testing.T) {
  logger, exportLogs := logtest.Captured(t)

  t.Run("test my thing", func(t *testing.T) {
    fooBar(logger)
  })

  // export log entries that were written during this time
  logs := exportLogs()
}
```
