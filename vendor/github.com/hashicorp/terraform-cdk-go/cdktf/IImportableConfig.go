// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

// Experimental.
type IImportableConfig interface {
	// Experimental.
	ImportId() *string
	// Experimental.
	SetImportId(i *string)
	// Experimental.
	Provider() TerraformProvider
	// Experimental.
	SetProvider(p TerraformProvider)
	// Experimental.
	TerraformResourceType() *string
	// Experimental.
	SetTerraformResourceType(t *string)
}

// The jsii proxy for IImportableConfig
type jsiiProxy_IImportableConfig struct {
	_ byte // padding
}

func (j *jsiiProxy_IImportableConfig) ImportId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"importId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_IImportableConfig)SetImportId(val *string) {
	if err := j.validateSetImportIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"importId",
		val,
	)
}

func (j *jsiiProxy_IImportableConfig) Provider() TerraformProvider {
	var returns TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_IImportableConfig)SetProvider(val TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_IImportableConfig) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_IImportableConfig)SetTerraformResourceType(val *string) {
	if err := j.validateSetTerraformResourceTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResourceType",
		val,
	)
}

