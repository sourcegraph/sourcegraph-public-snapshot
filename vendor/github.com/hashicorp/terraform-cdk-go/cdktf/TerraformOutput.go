// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"

	"github.com/aws/constructs-go/constructs/v10"
)

// Experimental.
type TerraformOutput interface {
	TerraformElement
	// Experimental.
	CdktfStack() TerraformStack
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	// Experimental.
	DependsOn() *[]ITerraformDependable
	// Experimental.
	SetDependsOn(val *[]ITerraformDependable)
	// Experimental.
	Description() *string
	// Experimental.
	SetDescription(val *string)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	// The tree node.
	// Experimental.
	Node() constructs.Node
	// Experimental.
	Precondition() *Precondition
	// Experimental.
	SetPrecondition(val *Precondition)
	// Experimental.
	RawOverrides() interface{}
	// Experimental.
	Sensitive() *bool
	// Experimental.
	SetSensitive(val *bool)
	// Experimental.
	StaticId() *bool
	// Experimental.
	SetStaticId(val *bool)
	// Experimental.
	Value() interface{}
	// Experimental.
	SetValue(val interface{})
	// Experimental.
	AddOverride(path *string, value interface{})
	// Overrides the auto-generated logical ID with a specific ID.
	// Experimental.
	OverrideLogicalId(newLogicalId *string)
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	// Experimental.
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	SynthesizeHclAttributes() *map[string]interface{}
	// Experimental.
	ToHclTerraform() interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	// Experimental.
	ToString() *string
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for TerraformOutput
type jsiiProxy_TerraformOutput struct {
	jsiiProxy_TerraformElement
}

func (j *jsiiProxy_TerraformOutput) CdktfStack() TerraformStack {
	var returns TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformOutput) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformOutput) DependsOn() *[]ITerraformDependable {
	var returns *[]ITerraformDependable
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformOutput) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformOutput) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformOutput) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformOutput) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformOutput) Precondition() *Precondition {
	var returns *Precondition
	_jsii_.Get(
		j,
		"precondition",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformOutput) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformOutput) Sensitive() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"sensitive",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformOutput) StaticId() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"staticId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformOutput) Value() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"value",
		&returns,
	)
	return returns
}


// Experimental.
func NewTerraformOutput(scope constructs.Construct, id *string, config *TerraformOutputConfig) TerraformOutput {
	_init_.Initialize()

	if err := validateNewTerraformOutputParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_TerraformOutput{}

	_jsii_.Create(
		"cdktf.TerraformOutput",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Experimental.
func NewTerraformOutput_Override(t TerraformOutput, scope constructs.Construct, id *string, config *TerraformOutputConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.TerraformOutput",
		[]interface{}{scope, id, config},
		t,
	)
}

func (j *jsiiProxy_TerraformOutput)SetDependsOn(val *[]ITerraformDependable) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_TerraformOutput)SetDescription(val *string) {
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_TerraformOutput)SetPrecondition(val *Precondition) {
	if err := j.validateSetPreconditionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"precondition",
		val,
	)
}

func (j *jsiiProxy_TerraformOutput)SetSensitive(val *bool) {
	_jsii_.Set(
		j,
		"sensitive",
		val,
	)
}

func (j *jsiiProxy_TerraformOutput)SetStaticId(val *bool) {
	if err := j.validateSetStaticIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"staticId",
		val,
	)
}

func (j *jsiiProxy_TerraformOutput)SetValue(val interface{}) {
	if err := j.validateSetValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"value",
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
func TerraformOutput_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateTerraformOutput_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.TerraformOutput",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func TerraformOutput_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateTerraformOutput_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.TerraformOutput",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func TerraformOutput_IsTerraformOutput(x interface{}) *bool {
	_init_.Initialize()

	if err := validateTerraformOutput_IsTerraformOutputParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.TerraformOutput",
		"isTerraformOutput",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformOutput) AddOverride(path *string, value interface{}) {
	if err := t.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		t,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (t *jsiiProxy_TerraformOutput) OverrideLogicalId(newLogicalId *string) {
	if err := t.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		t,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (t *jsiiProxy_TerraformOutput) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		t,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (t *jsiiProxy_TerraformOutput) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		t,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformOutput) SynthesizeHclAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		t,
		"synthesizeHclAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformOutput) ToHclTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		t,
		"toHclTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformOutput) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		t,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformOutput) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		t,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformOutput) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		t,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

