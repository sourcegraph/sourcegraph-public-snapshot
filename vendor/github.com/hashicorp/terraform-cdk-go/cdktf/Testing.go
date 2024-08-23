// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"

	"github.com/aws/constructs-go/constructs/v10"
)

// Testing utilities for cdktf applications.
// Experimental.
type Testing interface {
}

// The jsii proxy struct for Testing
type jsiiProxy_Testing struct {
	_ byte // padding
}

// Experimental.
func NewTesting() Testing {
	_init_.Initialize()

	j := jsiiProxy_Testing{}

	_jsii_.Create(
		"cdktf.Testing",
		nil, // no parameters
		&j,
	)

	return &j
}

// Experimental.
func NewTesting_Override(t Testing) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.Testing",
		nil, // no parameters
		t,
	)
}

// Returns an app for testing with the following properties: - Output directory is a temp dir.
// Experimental.
func Testing_App(options *TestingAppConfig) App {
	_init_.Initialize()

	if err := validateTesting_AppParameters(options); err != nil {
		panic(err)
	}
	var returns App

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"app",
		[]interface{}{options},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_EnableFutureFlags(app App) App {
	_init_.Initialize()

	if err := validateTesting_EnableFutureFlagsParameters(app); err != nil {
		panic(err)
	}
	var returns App

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"enableFutureFlags",
		[]interface{}{app},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_FakeCdktfJsonPath(app App) App {
	_init_.Initialize()

	if err := validateTesting_FakeCdktfJsonPathParameters(app); err != nil {
		panic(err)
	}
	var returns App

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"fakeCdktfJsonPath",
		[]interface{}{app},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_FullSynth(stack TerraformStack) *string {
	_init_.Initialize()

	if err := validateTesting_FullSynthParameters(stack); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"fullSynth",
		[]interface{}{stack},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_RenderConstructTree(construct constructs.IConstruct) *string {
	_init_.Initialize()

	if err := validateTesting_RenderConstructTreeParameters(construct); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"renderConstructTree",
		[]interface{}{construct},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_SetupJest() {
	_init_.Initialize()

	_jsii_.StaticInvokeVoid(
		"cdktf.Testing",
		"setupJest",
		nil, // no parameters
	)
}

// Experimental.
func Testing_StubVersion(app App) App {
	_init_.Initialize()

	if err := validateTesting_StubVersionParameters(app); err != nil {
		panic(err)
	}
	var returns App

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"stubVersion",
		[]interface{}{app},
		&returns,
	)

	return returns
}

// Returns the Terraform synthesized JSON.
// Experimental.
func Testing_Synth(stack TerraformStack, runValidations *bool) *string {
	_init_.Initialize()

	if err := validateTesting_SynthParameters(stack); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"synth",
		[]interface{}{stack, runValidations},
		&returns,
	)

	return returns
}

// Returns the Terraform synthesized JSON.
// Experimental.
func Testing_SynthHcl(stack TerraformStack, runValidations *bool, returnMetadata *bool) interface{} {
	_init_.Initialize()

	if err := validateTesting_SynthHclParameters(stack); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"synthHcl",
		[]interface{}{stack, runValidations, returnMetadata},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_SynthScope(fn IScopeCallback) *string {
	_init_.Initialize()

	if err := validateTesting_SynthScopeParameters(fn); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"synthScope",
		[]interface{}{fn},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_ToBeValidTerraform(received *string) *bool {
	_init_.Initialize()

	if err := validateTesting_ToBeValidTerraformParameters(received); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"toBeValidTerraform",
		[]interface{}{received},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_ToHaveDataSource(received *string, resourceType *string) *bool {
	_init_.Initialize()

	if err := validateTesting_ToHaveDataSourceParameters(received, resourceType); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"toHaveDataSource",
		[]interface{}{received, resourceType},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_ToHaveDataSourceWithProperties(received *string, resourceType *string, properties *map[string]interface{}) *bool {
	_init_.Initialize()

	if err := validateTesting_ToHaveDataSourceWithPropertiesParameters(received, resourceType); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"toHaveDataSourceWithProperties",
		[]interface{}{received, resourceType, properties},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_ToHaveProvider(received *string, resourceType *string) *bool {
	_init_.Initialize()

	if err := validateTesting_ToHaveProviderParameters(received, resourceType); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"toHaveProvider",
		[]interface{}{received, resourceType},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_ToHaveProviderWithProperties(received *string, resourceType *string, properties *map[string]interface{}) *bool {
	_init_.Initialize()

	if err := validateTesting_ToHaveProviderWithPropertiesParameters(received, resourceType); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"toHaveProviderWithProperties",
		[]interface{}{received, resourceType, properties},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_ToHaveResource(received *string, resourceType *string) *bool {
	_init_.Initialize()

	if err := validateTesting_ToHaveResourceParameters(received, resourceType); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"toHaveResource",
		[]interface{}{received, resourceType},
		&returns,
	)

	return returns
}

// Experimental.
func Testing_ToHaveResourceWithProperties(received *string, resourceType *string, properties *map[string]interface{}) *bool {
	_init_.Initialize()

	if err := validateTesting_ToHaveResourceWithPropertiesParameters(received, resourceType); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.Testing",
		"toHaveResourceWithProperties",
		[]interface{}{received, resourceType, properties},
		&returns,
	)

	return returns
}

