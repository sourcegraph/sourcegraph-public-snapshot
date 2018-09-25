# Changelog

## [Pending Release](https://github.com/lightstep/lightstep-tracer-go/compare/v0.15.0...HEAD)

## [v0.15.0](https://github.com/lightstep/lightstep-tracer-go/compare/v0.14.0...v0.15.0)
* We are replacing the internal diagnostic logging with a more flexible “Event” framework. This enables you to track and understand tracer problems with metrics and logging tools of your choice.
* We are also changed the types of the Close() and Flush() methods to take a context parameter to support cancellation. These changes are *not* backwards compatible and *you will need to update your instrumentation*. We are providing a NewTracerv0_14() method that is a drop-in replacement for the previous version.

## [v0.14.0](https://github.com/lightstep/lightstep-tracer-go/compare/v0.13.0...v0.14.0)
* Flush buffer syncronously on Close
* Flush twice if a flush is already in flight.
* remove gogo in favor of golang/protobuf
* requires grpc-go >= 1.4.0

## [v0.13.0](https://github.com/lightstep/lightstep-tracer-go/compare/v0.12.0...v0.13.0) 
* BasicTracer has been removed.
* Tracer now takes a SpanRecorder as an option.
* Tracer interface now includes Close and Flush.
* Tests redone with ginkgo/gomega.

## [v0.12.0](https://github.com/lightstep/lightstep-tracer-go/compare/v0.11.0...v0.12.0)
* Added CloseTracer function to flush and close a lightstep recorder.

## [v0.11.0](https://github.com/lightstep/lightstep-tracer-go/compare/v0.10.0...v0.11.0)
* Thrift transport is now deprecated, gRPC is the default.