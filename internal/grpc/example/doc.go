// Package example contains didactic examples for the basic usage / implementation of gRPC including:
//
// - All the basic Protobuf types (e.g. primitives, enums, messages, one-ofs, etc.)
// - All the basic RPC types (e.g. unary, server streaming, client streaming, bidirectional streaming)
// - Error handling (e.g. gRPC errors, wrapping status errors, etc.)
// - Implementing a gRPC server (with proper separation of concerns)
// - Implementing a gRPC client
// - Some known footguns (non-utf8 strings, huge messages, etc.)
// - Some Sourcegraph specific helper packages and patterns (grpc/defaults, grpc/streamio, etc.)
//
//	The service is a simple weather-reporting service that allows you to query the current weather, send updated weather
//	location for a given sensor, receive severe weather alerts and more.
//
//	The examples are organized into the following directories:
//	- weather/v1: contains the protobuf definitions for the weather service
//	- server/: contains the implementation of the gRPC server
//	- client/: contains the implementation of the gRPC client
//
//
//	When going through this example for the first time, it is recommended to:
//
//	1. Read the protobuf definitions in weather/v1/weather.proto to get a sense of the service.
//	2. Run the server and client examples in server/ and client/ (via server/run-server.sh and client/run-client.sh) respectively to see the service in action.
//	3. Read the implementation of the server and client to get a sense of how things are implemented, and follow the explanatory comments in the code.
package example
