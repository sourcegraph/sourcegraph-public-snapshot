// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package configopaque // import "go.opentelemetry.io/collector/config/configopaque"

import (
	"fmt"
)

// String alias that is marshaled and printed in an opaque way.
// To recover the original value, cast it to a string.
type String string

const maskedString = "[REDACTED]"

// MarshalText marshals the string as `[REDACTED]`.
func (s String) MarshalText() ([]byte, error) {
	return []byte(maskedString), nil
}

// String formats the string as `[REDACTED]`.
// This is used for the %s and %q verbs.
func (s String) String() string {
	return maskedString
}

// GoString formats the string as `[REDACTED]`.
// This is used for the %#v verb.
func (s String) GoString() string {
	return fmt.Sprintf("%#v", maskedString)
}

// MarshalBinary marshals the string `[REDACTED]` as []byte.
func (s String) MarshalBinary() (text []byte, err error) {
	return []byte(maskedString), nil
}
