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
	"context"
	"time"

	connect "connectrpc.com/connect"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	version             = "0.6.0-dev"
	semanticVersion     = "semver:" + version
	instrumentationName = "connectrpc.com/otelconnect"
	grpcProtocol        = "grpc"
	grpcwebString       = "grpcweb"
	grpcwebProtocol     = "grpc_web"
	connectString       = "connect"
	connectProtocol     = "connect_rpc"
)

type config struct {
	filter             func(context.Context, connect.Spec) bool
	filterAttribute    AttributeFilter
	meter              metric.Meter
	tracer             trace.Tracer
	propagator         propagation.TextMapPropagator
	now                func() time.Time
	trustRemote        bool
	requestHeaderKeys  []string
	responseHeaderKeys []string
	omitTraceEvents    bool
}
