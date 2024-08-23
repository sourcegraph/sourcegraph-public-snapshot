// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package experr // import "go.opentelemetry.io/collector/exporter/internal/experr"

import "errors"

type shutdownErr struct {
	err error
}

func NewShutdownErr(err error) error {
	return shutdownErr{err: err}
}

func (s shutdownErr) Error() string {
	return "interrupted due to shutdown: " + s.err.Error()
}

func (s shutdownErr) Unwrap() error {
	return s.err
}

func IsShutdownErr(err error) bool {
	var sdErr shutdownErr
	return errors.As(err, &sdErr)
}
