// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Expressions in connection blocks cannot refer to their parent resource by name.
//
// References create dependencies, and referring to a resource by name within its own block would create a dependency cycle.
// Instead, expressions can use the self object, which represents the connection's parent resource and has all of that resource's attributes.
// For example, use self.public_ip to reference an aws_instance's public_ip attribute.
// Experimental.
type TerraformSelf interface {
}

// The jsii proxy struct for TerraformSelf
type jsiiProxy_TerraformSelf struct {
	_ byte // padding
}

// Experimental.
func NewTerraformSelf() TerraformSelf {
	_init_.Initialize()

	j := jsiiProxy_TerraformSelf{}

	_jsii_.Create(
		"cdktf.TerraformSelf",
		nil, // no parameters
		&j,
	)

	return &j
}

// Experimental.
func NewTerraformSelf_Override(t TerraformSelf) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.TerraformSelf",
		nil, // no parameters
		t,
	)
}

// Only usable within a connection block to reference the connections parent resource.
//
// Access a property on the resource like this: `getAny("hostPort")`.
// Experimental.
func TerraformSelf_GetAny(key *string) interface{} {
	_init_.Initialize()

	if err := validateTerraformSelf_GetAnyParameters(key); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.TerraformSelf",
		"getAny",
		[]interface{}{key},
		&returns,
	)

	return returns
}

// Only usable within a connection block to reference the connections parent resource.
//
// Access a property on the resource like this: `getNumber("hostPort")`.
// Experimental.
func TerraformSelf_GetNumber(key *string) *float64 {
	_init_.Initialize()

	if err := validateTerraformSelf_GetNumberParameters(key); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.TerraformSelf",
		"getNumber",
		[]interface{}{key},
		&returns,
	)

	return returns
}

// Only usable within a connection block to reference the connections parent resource.
//
// Access a property on the resource like this: `getString("publicIp")`.
// Experimental.
func TerraformSelf_GetString(key *string) *string {
	_init_.Initialize()

	if err := validateTerraformSelf_GetStringParameters(key); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.TerraformSelf",
		"getString",
		[]interface{}{key},
		&returns,
	)

	return returns
}

