// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package consumererror // import "go.opentelemetry.io/collector/consumer/consumererror"

import "errors"

// permanent is an error that will be always returned if its source
// receives the same inputs.
type permanent struct {
	err error
}

// NewPermanent wraps an error to indicate that it is a permanent error, i.e. an
// error that will be always returned if its source receives the same inputs.
func NewPermanent(err error) error {
	return permanent{err: err}
}

func (p permanent) Error() string {
	return "Permanent error: " + p.err.Error()
}

// Unwrap returns the wrapped error for functions Is and As in standard package errors.
func (p permanent) Unwrap() error {
	return p.err
}

// IsPermanent checks if an error was wrapped with the NewPermanent function, which
// is used to indicate that a given error will always be returned in the case
// that its sources receives the same input.
func IsPermanent(err error) bool {
	if err == nil {
		return false
	}
	return errors.As(err, &permanent{})
}
