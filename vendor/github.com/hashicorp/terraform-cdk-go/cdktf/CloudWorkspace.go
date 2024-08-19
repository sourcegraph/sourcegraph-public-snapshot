// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// A cloud workspace can either be a single named workspace, or a list of tagged workspaces.
// Experimental.
type CloudWorkspace interface {
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for CloudWorkspace
type jsiiProxy_CloudWorkspace struct {
	_ byte // padding
}

// Experimental.
func NewCloudWorkspace_Override(c CloudWorkspace) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.CloudWorkspace",
		nil, // no parameters
		c,
	)
}

func (c *jsiiProxy_CloudWorkspace) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

