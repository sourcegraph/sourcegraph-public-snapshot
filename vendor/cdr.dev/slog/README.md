# slog

[![GitHub Release](https://img.shields.io/github/v/release/cdr/slog?color=6b9ded&sort=semver)](https://github.com/cdr/slog/releases)
[![GoDoc](https://godoc.org/cdr.dev/slog?status.svg)](https://godoc.org/cdr.dev/slog)
[![Coveralls](https://img.shields.io/coveralls/github/cdr/slog?color=65d6a4)](https://coveralls.io/github/cdr/slog)
[![CI Status](https://github.com/cdr/slog/workflows/ci/badge.svg)](https://github.com/cdr/slog/actions)

slog is a minimal structured logging library for Go.

## Install

```bash
go get cdr.dev/slog
```

## Features

- Minimal API
- First class [context.Context](https://blog.golang.org/context) support
- First class [testing.TB](https://godoc.org/cdr.dev/slog/sloggers/slogtest) support
  - Package [slogtest/assert](https://godoc.org/cdr.dev/slog/sloggers/slogtest/assert) provides test assertion helpers
- Beautiful human readable logging output
  - Prints multiline fields and errors nicely
- Machine readable JSON output
- [GCP Stackdriver](https://godoc.org/cdr.dev/slog/sloggers/slogstackdriver) support
- [Stdlib](https://godoc.org/cdr.dev/slog#Stdlib) log adapter
- Skip caller frames with [slog.Helper](https://godoc.org/cdr.dev/slog#Helper)
- Encodes values as if with `json.Marshal`
- Transparently log [opencensus](https://godoc.org/go.opencensus.io/trace) trace and span IDs
- [Single dependency](https://godoc.org/cdr.dev/slog?imports) on go.opencensus.io
- Log to multiple sinks

## Example

Many more examples available at [godoc](https://godoc.org/cdr.dev/slog#pkg-examples).

```go
log := slog.Make(sloghuman.Sink(os.Stdout))

log.Info(context.Background(), "my message here",
    slog.F("field_name", "something or the other"),
    slog.F("some_map", slog.M(
        slog.F("nested_fields", time.Date(2000, time.February, 5, 4, 4, 4, 0, time.UTC)),
    )),
    slog.Error(
        xerrors.Errorf("wrap1: %w",
            xerrors.Errorf("wrap2: %w",
                io.EOF,
            ),
        ),
    ),
)
```

![Example output screenshot](https://i.imgur.com/KGRmQFo.png)

## Why?

At [Coder](https://github.com/cdr) we’ve used Uber’s [zap](https://github.com/uber-go/zap) for several years.
It is a fantastic library for performance. Thanks Uber!

However we felt the API and developer experience could be improved.

Here is a list of reasons how we improved on zap with slog.

1. `slog` has a minimal API surface

   - Compare [slog](https://godoc.org/cdr.dev/slog) to [zap](https://godoc.org/go.uber.org/zap) and
     [zapcore](https://godoc.org/go.uber.org/zap/zapcore).
   - The sprawling API makes zap hard to understand, use and extend.

1. `slog` has a concise semi typed API

   - We found zap's fully typed API cumbersome. It does offer a
     [sugared API](https://godoc.org/go.uber.org/zap#hdr-Choosing_a_Logger)
     but it's too easy to pass an invalid fields list since there is no static type checking.
     Furthermore, it's harder to read as there is no syntax grouping for each key value pair.
   - We wanted an API that only accepted the equivalent of [zap.Any](https://godoc.org/go.uber.org/zap#Any)
     for every field. This is [slog.F](https://godoc.org/cdr.dev/slog#F).

1. [`sloghuman`](https://godoc.org/cdr.dev/slog/sloggers/sloghuman) uses a very human readable format

   - It colors distinct parts of each line to make it easier to scan logs. Even the JSON that represents
     the fields in each log is syntax highlighted so that is very easy to scan. See the screenshot above.
     - zap lacks appropriate colors for different levels and fields.
   - slog automatically prints one multiline field after the log to make errors and such much easier to read.
     - zap logs multiline fields and errors stack traces as JSON strings which made them unreadable in a terminal.
   - When logging to JSON, slog automatically converts a [`golang.org/x/xerrors`](https://golang.org/x/xerrors) chain
     into an array with fields for the location and wrapping messages.

1. Full [context.Context](https://blog.golang.org/context) support

   - `slog` lets you set fields in a `context.Context` such that any log with the context prints those fields.
   - We wanted to be able to pull up all relevant logs for a given trace, user or request. With zap, we were plugging
     these fields in for every relevant log or passing around a logger with the fields set. This became very verbose.

1. Simple and easy to extend

   - A new backend only has to implement the simple Sink interface.
   - The Logger type provides a nice API around Sink but also implements
     Sink to allow for composition.
   - zap is hard and confusing to extend. There are too many structures and configuration options.

1. Structured logging of Go structures with `json.Marshal`

   - Entire encoding process is documented on [godoc](https://godoc.org/cdr.dev/slog#Map.MarshalJSON).
   - With zap, We found ourselves often implementing zap's
     [ObjectMarshaler](https://godoc.org/go.uber.org/zap/zapcore#ObjectMarshaler) to log Go structures. This was
     verbose and most of the time we ended up only implementing `fmt.Stringer` and using `zap.Stringer` instead.

1. slog takes inspiration from Go's stdlib and implements [`slog.Helper`](https://godoc.org/cdr.dev/slog#Helper)
   which works just like [`t.Helper`](https://golang.org/pkg/testing/#T.Helper)

   - It marks the calling function as a helper and skips it when reporting location info.
   - We had many helper functions for logging but we wanted the line reported to be of the parent function.
     zap has an [API](https://godoc.org/go.uber.org/zap#AddCallerSkip) for this but it's verbose and requires
     passing the logger around explicitly.

1. Tight integration with stdlib's [`testing`](https://golang.org/pkg/testing) package
   - You can configure [`slogtest`](https://godoc.org/cdr.dev/slog/sloggers/slogtest) to exit on any ERROR logs
     and it has a global stateless API that takes a `testing.TB` so you do not need to create a logger first.
   - Test assertion helpers are provided in [slogtest/assert](https://godoc.org/cdr.dev/slog/sloggers/slogtest/assert).
   - zap has [zaptest](https://godoc.org/go.uber.org/zap/zaptest) but the API surface is large and doesn't
     integrate well. It does not support any of the features described above.
