// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Iterator for the Terraform count property.
// Experimental.
type TerraformCount interface {
	// Experimental.
	Index() *float64
	// Experimental.
	ToString() *string
	// Experimental.
	ToTerraform() *float64
}

// The jsii proxy struct for TerraformCount
type jsiiProxy_TerraformCount struct {
	_ byte // padding
}

func (j *jsiiProxy_TerraformCount) Index() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"index",
		&returns,
	)
	return returns
}


// Experimental.
func TerraformCount_IsTerraformCount(x interface{}) *bool {
	_init_.Initialize()

	if err := validateTerraformCount_IsTerraformCountParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.TerraformCount",
		"isTerraformCount",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func TerraformCount_Of(count *float64) TerraformCount {
	_init_.Initialize()

	if err := validateTerraformCount_OfParameters(count); err != nil {
		panic(err)
	}
	var returns TerraformCount

	_jsii_.StaticInvoke(
		"cdktf.TerraformCount",
		"of",
		[]interface{}{count},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformCount) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		t,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformCount) ToTerraform() *float64 {
	var returns *float64

	_jsii_.Invoke(
		t,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

