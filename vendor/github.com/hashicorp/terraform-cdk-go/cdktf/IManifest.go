// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

// Experimental.
type IManifest interface {
	// Experimental.
	Stacks() *map[string]*StackManifest
	// Experimental.
	Version() *string
}

// The jsii proxy for IManifest
type jsiiProxy_IManifest struct {
	_ byte // padding
}

func (j *jsiiProxy_IManifest) Stacks() *map[string]*StackManifest {
	var returns *map[string]*StackManifest
	_jsii_.Get(
		j,
		"stacks",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_IManifest) Version() *string {
	var returns *string
	_jsii_.Get(
		j,
		"version",
		&returns,
	)
	return returns
}

