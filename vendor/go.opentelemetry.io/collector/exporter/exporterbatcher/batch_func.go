// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporterbatcher // import "go.opentelemetry.io/collector/exporter/exporterbatcher"

import "context"

// BatchMergeFunc is a function that merges two requests into a single request.
// Do not mutate the requests passed to the function if error can be returned after mutation or if the exporter is
// marked as not mutable.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type BatchMergeFunc[T any] func(context.Context, T, T) (T, error)

// BatchMergeSplitFunc is a function that merge and/or splits one or two requests into multiple requests based on the
// configured limit provided in MaxSizeConfig.
// All the returned requests MUST have a number of items that does not exceed the maximum number of items.
// Size of the last returned request MUST be less or equal than the size of any other returned request.
// The original request MUST not be mutated if error is returned after mutation or if the exporter is
// marked as not mutable. The length of the returned slice MUST not be 0. The optionalReq argument can be nil,
// make sure to check it before using.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type BatchMergeSplitFunc[T any] func(ctx context.Context, cfg MaxSizeConfig, optionalReq T, req T) ([]T, error)
