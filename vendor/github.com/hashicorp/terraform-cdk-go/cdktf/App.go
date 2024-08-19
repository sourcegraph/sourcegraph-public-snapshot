// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf/internal"
)

// Represents a cdktf application.
// Experimental.
type App interface {
	constructs.Construct
	// Experimental.
	HclOutput() *bool
	// Experimental.
	Manifest() Manifest
	// The tree node.
	// Experimental.
	Node() constructs.Node
	// The output directory into which resources will be synthesized.
	// Experimental.
	Outdir() *string
	// Whether to skip backend validation during synthesis of the app.
	// Experimental.
	SkipBackendValidation() *bool
	// Whether to skip all validations during synthesis of the app.
	// Experimental.
	SkipValidation() *bool
	// The stack which will be synthesized.
	//
	// If not set, all stacks will be synthesized.
	// Experimental.
	TargetStackId() *string
	// Creates a reference from one stack to another, invoked on prepareStack since it creates extra resources.
	// Experimental.
	CrossStackReference(fromStack TerraformStack, toStack TerraformStack, identifier *string) *string
	// Synthesizes all resources to the output directory.
	// Experimental.
	Synth()
	// Returns a string representation of this construct.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for App
type jsiiProxy_App struct {
	internal.Type__constructsConstruct
}

func (j *jsiiProxy_App) HclOutput() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"hclOutput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_App) Manifest() Manifest {
	var returns Manifest
	_jsii_.Get(
		j,
		"manifest",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_App) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_App) Outdir() *string {
	var returns *string
	_jsii_.Get(
		j,
		"outdir",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_App) SkipBackendValidation() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"skipBackendValidation",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_App) SkipValidation() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"skipValidation",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_App) TargetStackId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"targetStackId",
		&returns,
	)
	return returns
}


// Defines an app.
// Experimental.
func NewApp(config *AppConfig) App {
	_init_.Initialize()

	if err := validateNewAppParameters(config); err != nil {
		panic(err)
	}
	j := jsiiProxy_App{}

	_jsii_.Create(
		"cdktf.App",
		[]interface{}{config},
		&j,
	)

	return &j
}

// Defines an app.
// Experimental.
func NewApp_Override(a App, config *AppConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.App",
		[]interface{}{config},
		a,
	)
}

// Experimental.
func App_IsApp(x interface{}) *bool {
	_init_.Initialize()

	if err := validateApp_IsAppParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.App",
		"isApp",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct`
// instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on
// disk are seen as independent, completely different libraries. As a
// consequence, the class `Construct` in each copy of the `constructs` library
// is seen as a different class, and an instance of one class will not test as
// `instanceof` the other class. `npm install` will not create installations
// like this, but users may manually symlink construct libraries together or
// use a monorepo tool: in those cases, multiple copies of the `constructs`
// library can be accidentally installed, and `instanceof` will behave
// unpredictably. It is safest to avoid using `instanceof`, and using
// this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
// Experimental.
func App_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateApp_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.App",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func App_Of(construct constructs.IConstruct) App {
	_init_.Initialize()

	if err := validateApp_OfParameters(construct); err != nil {
		panic(err)
	}
	var returns App

	_jsii_.StaticInvoke(
		"cdktf.App",
		"of",
		[]interface{}{construct},
		&returns,
	)

	return returns
}

func (a *jsiiProxy_App) CrossStackReference(fromStack TerraformStack, toStack TerraformStack, identifier *string) *string {
	if err := a.validateCrossStackReferenceParameters(fromStack, toStack, identifier); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		a,
		"crossStackReference",
		[]interface{}{fromStack, toStack, identifier},
		&returns,
	)

	return returns
}

func (a *jsiiProxy_App) Synth() {
	_jsii_.InvokeVoid(
		a,
		"synth",
		nil, // no parameters
	)
}

func (a *jsiiProxy_App) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		a,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

