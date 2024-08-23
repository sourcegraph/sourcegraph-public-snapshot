// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal // import "go.opentelemetry.io/collector/confmap/internal"

import "go.opentelemetry.io/collector/featuregate"

const StrictlyTypedInputID = "confmap.strictlyTypedInput"

var StrictlyTypedInputGate = featuregate.GlobalRegistry().MustRegister(StrictlyTypedInputID,
	featuregate.StageAlpha,
	featuregate.WithRegisterFromVersion("v0.103.0"),
	featuregate.WithRegisterDescription("Makes type casting rules during configuration unmarshaling stricter. See https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/rfcs/env-vars.md for more details."),
)
