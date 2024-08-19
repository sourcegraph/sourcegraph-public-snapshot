// Package log provides an OpenTelemetry-compatible logging library for Sourcegraph.
// Subpackages are available that provide adapters, testing utilities, and additional log
// sinks.
//
// It encourages certain practices by design, such as by not having a global logger
// available, disallowing the use of compile-time loggers, and re-exporting a limited
// number of field types.
//
// Learn more: https://docs.sourcegraph.com/dev/how-to/add_logging
package log
