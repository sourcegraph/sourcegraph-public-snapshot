// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporterhelper // import "go.opentelemetry.io/collector/exporter/exporterhelper"

import (
	"context"
)

// Request represents a single request that can be sent to an external endpoint.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type Request interface {
	// Export exports the request to an external endpoint.
	Export(ctx context.Context) error
	// ItemsCount returns a number of basic items in the request where item is the smallest piece of data that can be
	// sent. For example, for OTLP exporter, this value represents the number of spans,
	// metric data points or log records.
	ItemsCount() int
}

// RequestErrorHandler is an optional interface that can be implemented by Request to provide a way handle partial
// temporary failures. For example, if some items failed to process and can be retried, this interface allows to
// return a new Request that contains the items left to be sent. Otherwise, the original Request should be returned.
// If not implemented, the original Request will be returned assuming the error is applied to the whole Request.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type RequestErrorHandler interface {
	Request
	// OnError returns a new Request may contain the items left to be sent if some items failed to process and can be retried.
	// Otherwise, it should return the original Request.
	OnError(error) Request
}

// extractPartialRequest returns a new Request that may contain the items left to be sent
// if only some items failed to process and can be retried. Otherwise, it returns the original Request.
func extractPartialRequest(req Request, err error) Request {
	if errReq, ok := req.(RequestErrorHandler); ok {
		return errReq.OnError(err)
	}
	return req
}
