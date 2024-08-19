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

// Package otelconnect provides OpenTelemetry tracing and metrics for
// [connectrpc.com/connect] servers and clients.
// The specification followed was the [OpenTelemetry specification] with both the [rpc metrics specification] and [rpc trace specification] implemented.
//
// [OpenTelemetry specification]: https://github.com/open-telemetry/opentelemetry-specification
// [rpc metrics specification]: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/semantic_conventions/rpc-metrics.md
// [rpc trace specification]: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/rpc.md
package otelconnect
