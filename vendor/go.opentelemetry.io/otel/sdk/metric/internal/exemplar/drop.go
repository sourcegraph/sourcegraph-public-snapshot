// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exemplar // import "go.opentelemetry.io/otel/sdk/metric/internal/exemplar"

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// Drop returns a [Reservoir] that drops all measurements it is offered.
func Drop() Reservoir { return &dropRes{} }

type dropRes struct{}

// Offer does nothing, all measurements offered will be dropped.
func (r *dropRes) Offer(context.Context, time.Time, Value, []attribute.KeyValue) {}

// Collect resets dest. No exemplars will ever be returned.
func (r *dropRes) Collect(dest *[]Exemplar) {
	*dest = (*dest)[:0]
}
