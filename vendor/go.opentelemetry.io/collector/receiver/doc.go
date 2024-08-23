// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package receiver defines components that allows the collector to receive metrics, traces and logs.
//
// Receiver receives data from a source (either from a remote source via network
// or scrapes from a local host) and pushes the data to the pipelines it is attached
// to by calling the nextConsumer.Consume*() function.
//
// # Error Handling
//
// The nextConsumer.Consume*() function may return an error to indicate that the data
// was not accepted. There are 2 types of possible errors: Permanent and non-Permanent.
// The receiver must check the type of the error using IsPermanent() helper.
//
// If the error is Permanent, then the nextConsumer.Consume*() call should not be
// retried with the same data. This typically happens when the data cannot be
// serialized by the exporter that is attached to the pipeline or when the destination
// refuses the data because it cannot decode it. The receiver must indicate to
// the source from which it received the data that the received data was bad, if the
// receiving protocol allows to do that. In case of OTLP/HTTP for example, this means
// that HTTP 400 response is returned to the sender.
//
// If the error is non-Permanent then the nextConsumer.Consume*() call should be retried
// with the same data. This may be done by the receiver itself, however typically it is
// done by the original sender, after the receiver returns a response to the sender
// indicating that the Collector is currently overloaded and the request must be
// retried. In case of OTLP/HTTP for example, this means that HTTP 429 or 503 response
// is returned.
//
// # Acknowledgment and Checkpointing
//
// The receivers that receive data via a network protocol that support acknowledgments
// MUST follow this order of operations:
//   - Receive data from some sender (typically from a network).
//   - Push received data to the pipeline by calling nextConsumer.Consume*() function.
//   - Acknowledge successful data receipt to the sender if Consume*() succeeded or
//     return a failure to the sender if Consume*() returned an error.
//
// This ensures there are strong delivery guarantees once the data is acknowledged
// by the Collector.
//
// Similarly, receivers that use checkpointing to remember the position of last processed
// data (e.g. via storage extension) MUST store the checkpoint only AFTER the Consume*()
// call returns.
package receiver // import "go.opentelemetry.io/collector/receiver"
