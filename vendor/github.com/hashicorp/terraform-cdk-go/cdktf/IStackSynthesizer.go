// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

// Encodes information how a certain Stack should be deployed inspired by AWS CDK v2 implementation (synth functionality was removed in constructs v10).
// Experimental.
type IStackSynthesizer interface {
	// Synthesize the associated stack to the session.
	// Experimental.
	Synthesize(session ISynthesisSession)
}

// The jsii proxy for IStackSynthesizer
type jsiiProxy_IStackSynthesizer struct {
	_ byte // padding
}

func (i *jsiiProxy_IStackSynthesizer) Synthesize(session ISynthesisSession) {
	if err := i.validateSynthesizeParameters(session); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		i,
		"synthesize",
		[]interface{}{session},
	)
}

