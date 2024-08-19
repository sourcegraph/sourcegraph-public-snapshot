// Copyright 2022-2023 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelconnect

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

const (
	metricKeyFormat     = "rpc.%s.%s"
	durationKey         = "duration"
	durationDesc        = "Measures the duration of inbound RPC."
	requestSizeKey      = "request.size"
	requestSizeDesc     = "Measures size of RPC request messages (uncompressed)."
	responseSizeKey     = "response.size"
	responseSizeDesc    = "Measures size of RPC response messages (uncompressed)."
	requestsPerRPCKey   = "requests_per_rpc"
	requestsPerRPCDesc  = "Measures the number of messages received per RPC. Should be 1 for all non-streaming RPCs."
	responsesPerRPCKey  = "responses_per_rpc"
	responsesPerRPCDesc = "Measures the number of messages received per RPC. Should be 1 for all non-streaming RPCs."
	messageKey          = "message"
	serverKey           = "server"
	clientKey           = "client"
	requestKey          = "request"
	responseKey         = "response"
	unitDimensionless   = "1"
	unitBytes           = "By"
	unitMilliseconds    = "ms"
)

type instruments struct {
	duration        metric.Int64Histogram
	requestSize     metric.Int64Histogram
	responseSize    metric.Int64Histogram
	requestsPerRPC  metric.Int64Histogram
	responsesPerRPC metric.Int64Histogram
}

// createInstruments creates the metrics for the interceptor.
func createInstruments(meter metric.Meter, interceptorType string) (instruments, error) {
	duration, err := meter.Int64Histogram(
		formatkeys(interceptorType, durationKey),
		metric.WithUnit(unitMilliseconds),
		metric.WithDescription(durationDesc),
	)
	if err != nil {
		return instruments{}, err
	}
	requestSize, err := meter.Int64Histogram(
		formatkeys(interceptorType, requestSizeKey),
		metric.WithUnit(unitBytes),
		metric.WithDescription(requestSizeDesc),
	)
	if err != nil {
		return instruments{}, err
	}
	responseSize, err := meter.Int64Histogram(
		formatkeys(interceptorType, responseSizeKey),
		metric.WithUnit(unitBytes),
		metric.WithDescription(responseSizeDesc),
	)
	if err != nil {
		return instruments{}, err
	}
	requestsPerRPC, err := meter.Int64Histogram(
		formatkeys(interceptorType, requestsPerRPCKey),
		metric.WithUnit(unitDimensionless),
		metric.WithDescription(requestsPerRPCDesc),
	)
	if err != nil {
		return instruments{}, err
	}
	responsesPerRPC, err := meter.Int64Histogram(
		formatkeys(interceptorType, responsesPerRPCKey),
		metric.WithUnit(unitDimensionless),
		metric.WithDescription(responsesPerRPCDesc),
	)
	if err != nil {
		return instruments{}, err
	}
	return instruments{
		duration:        duration,
		requestSize:     requestSize,
		responseSize:    responseSize,
		requestsPerRPC:  requestsPerRPC,
		responsesPerRPC: responsesPerRPC,
	}, nil
}

func formatkeys(interceptorType string, metricName string) string {
	return fmt.Sprintf(metricKeyFormat, interceptorType, metricName)
}
