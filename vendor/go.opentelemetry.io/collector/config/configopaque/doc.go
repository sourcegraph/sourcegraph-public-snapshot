// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package configopaque implements a String type alias to mask sensitive information.
// Use configopaque.String on the type of sensitive fields, to mask the
// opaque string as `[REDACTED]`.
//
// This ensures that no sensitive information is leaked in logs or when printing the
// full Collector configurations.
//
// The only way to view the value stored in a configopaque.String is to first convert
// it to a string by casting with the builtin `string` function.
//
// To achieve this, configopaque.String implements standard library interfaces
// like fmt.Stringer, encoding.TextMarshaler and others to ensure that the
// underlying value is masked when printed or serialized.
//
// If new interfaces that would leak opaque values are added to the standard library
// or become widely used in the Go ecosystem, these will eventually be implemented
// by configopaque.String as well. This is not considered a breaking change.
package configopaque // import "go.opentelemetry.io/collector/config/configopaque"
