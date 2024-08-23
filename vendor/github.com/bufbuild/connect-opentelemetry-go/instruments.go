// Copyright 2022-2023 Buf Technologies, Inc.
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
	"sync"

	"go.opentelemetry.io/otel/metric"
)

const (
	metricKeyFormat    = "rpc.%s.%s"
	durationKey        = "duration"
	requestSizeKey     = "request.size"
	responseSizeKey    = "response.size"
	requestsPerRPCKey  = "requests_per_rpc"
	responsesPerRPCKey = "responses_per_rpc"
	messageKey         = "message"
	serverKey          = "server"
	clientKey          = "client"
	requestKey         = "request"
	responseKey        = "response"
	unitDimensionless  = "1"
	unitBytes          = "By"
	unitMilliseconds   = "ms"
)

type instruments struct {
	initOnce        sync.Once
	initErr         error
	duration        metric.Int64Histogram
	requestSize     metric.Int64Histogram
	responseSize    metric.Int64Histogram
	requestsPerRPC  metric.Int64Histogram
	responsesPerRPC metric.Int64Histogram
}

func (i *instruments) init(meter metric.Meter, isClient bool) {
	i.initOnce.Do(func() {
		interceptorType := serverKey
		if isClient {
			interceptorType = clientKey
		}
		i.duration, i.initErr = meter.Int64Histogram(
			formatkeys(interceptorType, durationKey),
			metric.WithUnit(unitMilliseconds),
		)
		if i.initErr != nil {
			return
		}
		i.requestSize, i.initErr = meter.Int64Histogram(
			formatkeys(interceptorType, requestSizeKey),
			metric.WithUnit(unitBytes),
		)
		if i.initErr != nil {
			return
		}
		i.responseSize, i.initErr = meter.Int64Histogram(
			formatkeys(interceptorType, responseSizeKey),
			metric.WithUnit(unitBytes),
		)
		if i.initErr != nil {
			return
		}
		i.requestsPerRPC, i.initErr = meter.Int64Histogram(
			formatkeys(interceptorType, requestsPerRPCKey),
			metric.WithUnit(unitDimensionless),
		)
		if i.initErr != nil {
			return
		}
		i.responsesPerRPC, i.initErr = meter.Int64Histogram(
			formatkeys(interceptorType, responsesPerRPCKey),
			metric.WithUnit(unitDimensionless),
		)
	})
}

func formatkeys(interceptorType string, metricName string) string {
	return fmt.Sprintf(metricKeyFormat, interceptorType, metricName)
}
