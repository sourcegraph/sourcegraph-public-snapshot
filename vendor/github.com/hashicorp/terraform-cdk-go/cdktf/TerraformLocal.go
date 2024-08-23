// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"

	"github.com/aws/constructs-go/constructs/v10"
)

// Experimental.
type TerraformLocal interface {
	TerraformElement
	ITerraformAddressable
	// Experimental.
	AsAnyMap() *map[string]interface{}
	// Experimental.
	AsBoolean() IResolvable
	// Experimental.
	AsBooleanMap() *map[string]*bool
	// Experimental.
	AsList() *[]*string
	// Experimental.
	AsNumber() *float64
	// Experimental.
	AsNumberMap() *map[string]*float64
	// Experimental.
	AsString() *string
	// Experimental.
	AsStringMap() *map[string]*string
	// Experimental.
	CdktfStack() TerraformStack
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	// Experimental.
	Expression() interface{}
	// Experimental.
	SetExpression(val interface{})
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	// The tree node.
	// Experimental.
	Node() constructs.Node
	// Experimental.
	RawOverrides() interface{}
	// Experimental.
	AddOverride(path *string, value interface{})
	// Overrides the auto-generated logical ID with a specific ID.
	// Experimental.
	OverrideLogicalId(newLogicalId *string)
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	// Experimental.
	ToHclTerraform() interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	//
	// Returns: a string token referencing the value of this local.
	// Experimental.
	ToString() *string
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for TerraformLocal
type jsiiProxy_TerraformLocal struct {
	jsiiProxy_TerraformElement
	jsiiProxy_ITerraformAddressable
}

func (j *jsiiProxy_TerraformLocal) AsAnyMap() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"asAnyMap",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) AsBoolean() IResolvable {
	var returns IResolvable
	_jsii_.Get(
		j,
		"asBoolean",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) AsBooleanMap() *map[string]*bool {
	var returns *map[string]*bool
	_jsii_.Get(
		j,
		"asBooleanMap",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) AsList() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"asList",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) AsNumber() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"asNumber",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) AsNumberMap() *map[string]*float64 {
	var returns *map[string]*float64
	_jsii_.Get(
		j,
		"asNumberMap",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) AsString() *string {
	var returns *string
	_jsii_.Get(
		j,
		"asString",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) AsStringMap() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"asStringMap",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) CdktfStack() TerraformStack {
	var returns TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) Expression() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"expression",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformLocal) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}


// Experimental.
func NewTerraformLocal(scope constructs.Construct, id *string, expression interface{}) TerraformLocal {
	_init_.Initialize()

	if err := validateNewTerraformLocalParameters(scope, id, expression); err != nil {
		panic(err)
	}
	j := jsiiProxy_TerraformLocal{}

	_jsii_.Create(
		"cdktf.TerraformLocal",
		[]interface{}{scope, id, expression},
		&j,
	)

	return &j
}

// Experimental.
func NewTerraformLocal_Override(t TerraformLocal, scope constructs.Construct, id *string, expression interface{}) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.TerraformLocal",
		[]interface{}{scope, id, expression},
		t,
	)
}

func (j *jsiiProxy_TerraformLocal)SetExpression(val interface{}) {
	if err := j.validateSetExpressionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"expression",
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
func TerraformLocal_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateTerraformLocal_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.TerraformLocal",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func TerraformLocal_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateTerraformLocal_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.TerraformLocal",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformLocal) AddOverride(path *string, value interface{}) {
	if err := t.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		t,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (t *jsiiProxy_TerraformLocal) OverrideLogicalId(newLogicalId *string) {
	if err := t.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		t,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (t *jsiiProxy_TerraformLocal) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		t,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (t *jsiiProxy_TerraformLocal) ToHclTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		t,
		"toHclTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformLocal) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		t,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformLocal) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		t,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformLocal) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		t,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

