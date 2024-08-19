// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Experimental.
type PrefixedRemoteWorkspaces interface {
	IRemoteWorkspace
	// Experimental.
	Prefix() *string
	// Experimental.
	SetPrefix(val *string)
}

// The jsii proxy struct for PrefixedRemoteWorkspaces
type jsiiProxy_PrefixedRemoteWorkspaces struct {
	jsiiProxy_IRemoteWorkspace
}

func (j *jsiiProxy_PrefixedRemoteWorkspaces) Prefix() *string {
	var returns *string
	_jsii_.Get(
		j,
		"prefix",
		&returns,
	)
	return returns
}


// Experimental.
func NewPrefixedRemoteWorkspaces(prefix *string) PrefixedRemoteWorkspaces {
	_init_.Initialize()

	if err := validateNewPrefixedRemoteWorkspacesParameters(prefix); err != nil {
		panic(err)
	}
	j := jsiiProxy_PrefixedRemoteWorkspaces{}

	_jsii_.Create(
		"cdktf.PrefixedRemoteWorkspaces",
		[]interface{}{prefix},
		&j,
	)

	return &j
}

// Experimental.
func NewPrefixedRemoteWorkspaces_Override(p PrefixedRemoteWorkspaces, prefix *string) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.PrefixedRemoteWorkspaces",
		[]interface{}{prefix},
		p,
	)
}

func (j *jsiiProxy_PrefixedRemoteWorkspaces)SetPrefix(val *string) {
	if err := j.validateSetPrefixParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"prefix",
		val,
	)
}

