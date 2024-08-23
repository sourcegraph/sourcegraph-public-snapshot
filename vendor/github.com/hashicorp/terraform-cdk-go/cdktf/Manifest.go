// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Experimental.
type Manifest interface {
	IManifest
	// Experimental.
	HclOutput() *bool
	// Experimental.
	Outdir() *string
	// Experimental.
	StackFileName() *string
	// Experimental.
	Stacks() *map[string]*StackManifest
	// Experimental.
	Version() *string
	// Experimental.
	BuildManifest() IManifest
	// Experimental.
	ForStack(stack TerraformStack) *StackManifest
	// Experimental.
	WriteToFile()
}

// The jsii proxy struct for Manifest
type jsiiProxy_Manifest struct {
	jsiiProxy_IManifest
}

func (j *jsiiProxy_Manifest) HclOutput() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"hclOutput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Manifest) Outdir() *string {
	var returns *string
	_jsii_.Get(
		j,
		"outdir",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Manifest) StackFileName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"stackFileName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Manifest) Stacks() *map[string]*StackManifest {
	var returns *map[string]*StackManifest
	_jsii_.Get(
		j,
		"stacks",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Manifest) Version() *string {
	var returns *string
	_jsii_.Get(
		j,
		"version",
		&returns,
	)
	return returns
}


// Experimental.
func NewManifest(version *string, outdir *string, hclOutput *bool) Manifest {
	_init_.Initialize()

	if err := validateNewManifestParameters(version, outdir, hclOutput); err != nil {
		panic(err)
	}
	j := jsiiProxy_Manifest{}

	_jsii_.Create(
		"cdktf.Manifest",
		[]interface{}{version, outdir, hclOutput},
		&j,
	)

	return &j
}

// Experimental.
func NewManifest_Override(m Manifest, version *string, outdir *string, hclOutput *bool) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.Manifest",
		[]interface{}{version, outdir, hclOutput},
		m,
	)
}

func Manifest_FileName() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.Manifest",
		"fileName",
		&returns,
	)
	return returns
}

func Manifest_StackMetadataPath() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.Manifest",
		"stackMetadataPath",
		&returns,
	)
	return returns
}

func Manifest_StacksFolder() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.Manifest",
		"stacksFolder",
		&returns,
	)
	return returns
}

func (m *jsiiProxy_Manifest) BuildManifest() IManifest {
	var returns IManifest

	_jsii_.Invoke(
		m,
		"buildManifest",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_Manifest) ForStack(stack TerraformStack) *StackManifest {
	if err := m.validateForStackParameters(stack); err != nil {
		panic(err)
	}
	var returns *StackManifest

	_jsii_.Invoke(
		m,
		"forStack",
		[]interface{}{stack},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_Manifest) WriteToFile() {
	_jsii_.InvokeVoid(
		m,
		"writeToFile",
		nil, // no parameters
	)
}

