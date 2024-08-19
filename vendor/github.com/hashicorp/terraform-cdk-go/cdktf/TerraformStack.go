// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf/internal"
)

// Experimental.
type TerraformStack interface {
	constructs.Construct
	// Experimental.
	Dependencies() *[]TerraformStack
	// Experimental.
	SetDependencies(val *[]TerraformStack)
	// Experimental.
	MoveTargets() TerraformResourceTargets
	// Experimental.
	SetMoveTargets(val TerraformResourceTargets)
	// The tree node.
	// Experimental.
	Node() constructs.Node
	// Experimental.
	Synthesizer() IStackSynthesizer
	// Experimental.
	SetSynthesizer(val IStackSynthesizer)
	// Experimental.
	AddDependency(dependency TerraformStack)
	// Experimental.
	AddOverride(path *string, value interface{})
	// Returns the naming scheme used to allocate logical IDs.
	//
	// By default, uses
	// the `HashedAddressingScheme` but this method can be overridden to customize
	// this behavior.
	// Experimental.
	AllocateLogicalId(tfElement interface{}) *string
	// Experimental.
	AllProviders() *[]TerraformProvider
	// Experimental.
	DependsOn(stack TerraformStack) *bool
	// Experimental.
	EnsureBackendExists() TerraformBackend
	// Experimental.
	GetLogicalId(tfElement interface{}) *string
	// Experimental.
	HasResourceMove() *bool
	// Experimental.
	PrepareStack()
	// Experimental.
	RegisterIncomingCrossStackReference(fromStack TerraformStack) TerraformRemoteState
	// Experimental.
	RegisterOutgoingCrossStackReference(identifier *string) TerraformOutput
	// Run all validations on the stack.
	// Experimental.
	RunAllValidations()
	// Experimental.
	ToHclTerraform() *map[string]interface{}
	// Returns a string representation of this construct.
	// Experimental.
	ToString() *string
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for TerraformStack
type jsiiProxy_TerraformStack struct {
	internal.Type__constructsConstruct
}

func (j *jsiiProxy_TerraformStack) Dependencies() *[]TerraformStack {
	var returns *[]TerraformStack
	_jsii_.Get(
		j,
		"dependencies",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformStack) MoveTargets() TerraformResourceTargets {
	var returns TerraformResourceTargets
	_jsii_.Get(
		j,
		"moveTargets",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformStack) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformStack) Synthesizer() IStackSynthesizer {
	var returns IStackSynthesizer
	_jsii_.Get(
		j,
		"synthesizer",
		&returns,
	)
	return returns
}


// Experimental.
func NewTerraformStack(scope constructs.Construct, id *string) TerraformStack {
	_init_.Initialize()

	if err := validateNewTerraformStackParameters(scope, id); err != nil {
		panic(err)
	}
	j := jsiiProxy_TerraformStack{}

	_jsii_.Create(
		"cdktf.TerraformStack",
		[]interface{}{scope, id},
		&j,
	)

	return &j
}

// Experimental.
func NewTerraformStack_Override(t TerraformStack, scope constructs.Construct, id *string) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.TerraformStack",
		[]interface{}{scope, id},
		t,
	)
}

func (j *jsiiProxy_TerraformStack)SetDependencies(val *[]TerraformStack) {
	if err := j.validateSetDependenciesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"dependencies",
		val,
	)
}

func (j *jsiiProxy_TerraformStack)SetMoveTargets(val TerraformResourceTargets) {
	if err := j.validateSetMoveTargetsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"moveTargets",
		val,
	)
}

func (j *jsiiProxy_TerraformStack)SetSynthesizer(val IStackSynthesizer) {
	if err := j.validateSetSynthesizerParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"synthesizer",
		val,
	)
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
func TerraformStack_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateTerraformStack_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.TerraformStack",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func TerraformStack_IsStack(x interface{}) *bool {
	_init_.Initialize()

	if err := validateTerraformStack_IsStackParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.TerraformStack",
		"isStack",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func TerraformStack_Of(construct constructs.IConstruct) TerraformStack {
	_init_.Initialize()

	if err := validateTerraformStack_OfParameters(construct); err != nil {
		panic(err)
	}
	var returns TerraformStack

	_jsii_.StaticInvoke(
		"cdktf.TerraformStack",
		"of",
		[]interface{}{construct},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformStack) AddDependency(dependency TerraformStack) {
	if err := t.validateAddDependencyParameters(dependency); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		t,
		"addDependency",
		[]interface{}{dependency},
	)
}

func (t *jsiiProxy_TerraformStack) AddOverride(path *string, value interface{}) {
	if err := t.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		t,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (t *jsiiProxy_TerraformStack) AllocateLogicalId(tfElement interface{}) *string {
	if err := t.validateAllocateLogicalIdParameters(tfElement); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		t,
		"allocateLogicalId",
		[]interface{}{tfElement},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformStack) AllProviders() *[]TerraformProvider {
	var returns *[]TerraformProvider

	_jsii_.Invoke(
		t,
		"allProviders",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformStack) DependsOn(stack TerraformStack) *bool {
	if err := t.validateDependsOnParameters(stack); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.Invoke(
		t,
		"dependsOn",
		[]interface{}{stack},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformStack) EnsureBackendExists() TerraformBackend {
	var returns TerraformBackend

	_jsii_.Invoke(
		t,
		"ensureBackendExists",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformStack) GetLogicalId(tfElement interface{}) *string {
	if err := t.validateGetLogicalIdParameters(tfElement); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		t,
		"getLogicalId",
		[]interface{}{tfElement},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformStack) HasResourceMove() *bool {
	var returns *bool

	_jsii_.Invoke(
		t,
		"hasResourceMove",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformStack) PrepareStack() {
	_jsii_.InvokeVoid(
		t,
		"prepareStack",
		nil, // no parameters
	)
}

func (t *jsiiProxy_TerraformStack) RegisterIncomingCrossStackReference(fromStack TerraformStack) TerraformRemoteState {
	if err := t.validateRegisterIncomingCrossStackReferenceParameters(fromStack); err != nil {
		panic(err)
	}
	var returns TerraformRemoteState

	_jsii_.Invoke(
		t,
		"registerIncomingCrossStackReference",
		[]interface{}{fromStack},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformStack) RegisterOutgoingCrossStackReference(identifier *string) TerraformOutput {
	if err := t.validateRegisterOutgoingCrossStackReferenceParameters(identifier); err != nil {
		panic(err)
	}
	var returns TerraformOutput

	_jsii_.Invoke(
		t,
		"registerOutgoingCrossStackReference",
		[]interface{}{identifier},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformStack) RunAllValidations() {
	_jsii_.InvokeVoid(
		t,
		"runAllValidations",
		nil, // no parameters
	)
}

func (t *jsiiProxy_TerraformStack) ToHclTerraform() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		t,
		"toHclTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformStack) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		t,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformStack) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		t,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

