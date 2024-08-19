// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package consumer // import "go.opentelemetry.io/collector/consumer"

import (
	"errors"
)

// Capabilities describes the capabilities of a Processor.
type Capabilities struct {
	// MutatesData is set to true if Consume* function of the
	// processor modifies the input Traces, Logs or Metrics argument.
	// Processors which modify the input data MUST set this flag to true. If the processor
	// does not modify the data it MUST set this flag to false. If the processor creates
	// a copy of the data before modifying then this flag can be safely set to false.
	MutatesData bool
}

type baseConsumer interface {
	Capabilities() Capabilities
}

var errNilFunc = errors.New("nil consumer func")

type baseImpl struct {
	capabilities Capabilities
}

// Option to construct new consumers.
type Option func(*baseImpl)

// WithCapabilities overrides the default GetCapabilities function for a processor.
// The default GetCapabilities function returns mutable capabilities.
func WithCapabilities(capabilities Capabilities) Option {
	return func(o *baseImpl) {
		o.capabilities = capabilities
	}
}

// Capabilities returns the capabilities of the component
func (bs baseImpl) Capabilities() Capabilities {
	return bs.capabilities
}

func newBaseImpl(options ...Option) *baseImpl {
	bs := &baseImpl{
		capabilities: Capabilities{MutatesData: false},
	}

	for _, op := range options {
		op(bs)
	}

	return bs
}
