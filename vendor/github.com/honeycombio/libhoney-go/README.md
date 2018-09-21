# libhoney [![Build Status](https://travis-ci.org/honeycombio/libhoney-go.svg?branch=master)](https://travis-ci.org/honeycombio/libhoney-go)

Go library for sending events to [Honeycomb](https://honeycomb.io). (For more information, see the [documentation](https://honeycomb.io/docs/) and [Go SDK guide](https://honeycomb.io/docs/connect/go).)

## Installation:

```
go get -v github.com/honeycombio/libhoney-go
```

## Documentation

A godoc API reference is available at https://godoc.org/github.com/honeycombio/libhoney-go

## Example

Honeycomb can calculate all sorts of statistics, so send the values you care about and let us crunch the averages, percentiles, lower/upper bounds, cardinality -- whatever you want -- for you.

```go
import "github.com/honeycombio/libhoney-go"

// Call Init to configure libhoney
libhoney.Init(libhoney.Config{
  WriteKey: "YOUR_WRITE_KEY",
  Dataset: "honeycomb-golang-example",
})
defer libhoney.Close() // Flush any pending calls to Honeycomb

libhoney.SendNow(map[string]interface{}{
  "duration_ms": 153.12,
  "method": "get",
  "hostname": "appserver15",
  "payload_length": 27,
})
```

See the [`examples` directory](examples/read_json_log.go) for sample code demonstrating how to use events,
builders, fields, and dynamic fields.

## Contributions

Features, bug fixes and other changes to libhoney are gladly accepted. Please
open issues or a pull request with your change. Remember to add your name to the
CONTRIBUTORS file!

All contributions will be released under the Apache License 2.0.

