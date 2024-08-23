// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

// Represents a single session of synthesis.
//
// Passed into `TerraformStack.onSynthesize()` methods.
// originally from aws/constructs lib v3.3.126 (synth functionality was removed in constructs v10)
// Experimental.
type ISynthesisSession interface {
	// Experimental.
	Manifest() Manifest
	// The output directory for this synthesis session.
	// Experimental.
	Outdir() *string
	// Experimental.
	SkipValidation() *bool
}

// The jsii proxy for ISynthesisSession
type jsiiProxy_ISynthesisSession struct {
	_ byte // padding
}

func (j *jsiiProxy_ISynthesisSession) Manifest() Manifest {
	var returns Manifest
	_jsii_.Get(
		j,
		"manifest",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ISynthesisSession) Outdir() *string {
	var returns *string
	_jsii_.Get(
		j,
		"outdir",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ISynthesisSession) SkipValidation() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"skipValidation",
		&returns,
	)
	return returns
}

