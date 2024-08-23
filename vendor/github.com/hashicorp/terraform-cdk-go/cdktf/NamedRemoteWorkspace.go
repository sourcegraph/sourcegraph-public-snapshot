// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Experimental.
type NamedRemoteWorkspace interface {
	IRemoteWorkspace
	// Experimental.
	Name() *string
	// Experimental.
	SetName(val *string)
}

// The jsii proxy struct for NamedRemoteWorkspace
type jsiiProxy_NamedRemoteWorkspace struct {
	jsiiProxy_IRemoteWorkspace
}

func (j *jsiiProxy_NamedRemoteWorkspace) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}


// Experimental.
func NewNamedRemoteWorkspace(name *string) NamedRemoteWorkspace {
	_init_.Initialize()

	if err := validateNewNamedRemoteWorkspaceParameters(name); err != nil {
		panic(err)
	}
	j := jsiiProxy_NamedRemoteWorkspace{}

	_jsii_.Create(
		"cdktf.NamedRemoteWorkspace",
		[]interface{}{name},
		&j,
	)

	return &j
}

// Experimental.
func NewNamedRemoteWorkspace_Override(n NamedRemoteWorkspace, name *string) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.NamedRemoteWorkspace",
		[]interface{}{name},
		n,
	)
}

func (j *jsiiProxy_NamedRemoteWorkspace)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

