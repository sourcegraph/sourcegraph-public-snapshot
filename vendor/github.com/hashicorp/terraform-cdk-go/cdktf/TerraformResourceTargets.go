// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Experimental.
type TerraformResourceTargets interface {
	// Experimental.
	AddResourceTarget(resource TerraformResource, target *string)
	// Experimental.
	GetResourceByTarget(target *string) TerraformResource
}

// The jsii proxy struct for TerraformResourceTargets
type jsiiProxy_TerraformResourceTargets struct {
	_ byte // padding
}

// Experimental.
func NewTerraformResourceTargets() TerraformResourceTargets {
	_init_.Initialize()

	j := jsiiProxy_TerraformResourceTargets{}

	_jsii_.Create(
		"cdktf.TerraformResourceTargets",
		nil, // no parameters
		&j,
	)

	return &j
}

// Experimental.
func NewTerraformResourceTargets_Override(t TerraformResourceTargets) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.TerraformResourceTargets",
		nil, // no parameters
		t,
	)
}

func (t *jsiiProxy_TerraformResourceTargets) AddResourceTarget(resource TerraformResource, target *string) {
	if err := t.validateAddResourceTargetParameters(resource, target); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		t,
		"addResourceTarget",
		[]interface{}{resource, target},
	)
}

func (t *jsiiProxy_TerraformResourceTargets) GetResourceByTarget(target *string) TerraformResource {
	if err := t.validateGetResourceByTargetParameters(target); err != nil {
		panic(err)
	}
	var returns TerraformResource

	_jsii_.Invoke(
		t,
		"getResourceByTarget",
		[]interface{}{target},
		&returns,
	)

	return returns
}

