// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package configgrpc // import "go.opentelemetry.io/collector/config/configgrpc"

import (
	// Import the gzip package which auto-registers the gzip gRPC compressor.
	_ "google.golang.org/grpc/encoding/gzip"
)
