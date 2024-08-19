// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"

	"github.com/aws/constructs-go/constructs/v10"
)

// Experimental.
type DataTerraformRemoteStateConsul interface {
	TerraformRemoteState
	// Experimental.
	CdktfStack() TerraformStack
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
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
	// Experimental.
	Get(output *string) IResolvable
	// Experimental.
	GetBoolean(output *string) IResolvable
	// Experimental.
	GetList(output *string) *[]*string
	// Experimental.
	GetNumber(output *string) *float64
	// Experimental.
	GetString(output *string) *string
	// Overrides the auto-generated logical ID with a specific ID.
	// Experimental.
	OverrideLogicalId(newLogicalId *string)
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToHclTerraform() interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	// Experimental.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for DataTerraformRemoteStateConsul
type jsiiProxy_DataTerraformRemoteStateConsul struct {
	jsiiProxy_TerraformRemoteState
}

func (j *jsiiProxy_DataTerraformRemoteStateConsul) CdktfStack() TerraformStack {
	var returns TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DataTerraformRemoteStateConsul) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DataTerraformRemoteStateConsul) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DataTerraformRemoteStateConsul) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DataTerraformRemoteStateConsul) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_DataTerraformRemoteStateConsul) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}


// Experimental.
func NewDataTerraformRemoteStateConsul(scope constructs.Construct, id *string, config *DataTerraformRemoteStateConsulConfig) DataTerraformRemoteStateConsul {
	_init_.Initialize()

	if err := validateNewDataTerraformRemoteStateConsulParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_DataTerraformRemoteStateConsul{}

	_jsii_.Create(
		"cdktf.DataTerraformRemoteStateConsul",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Experimental.
func NewDataTerraformRemoteStateConsul_Override(d DataTerraformRemoteStateConsul, scope constructs.Construct, id *string, config *DataTerraformRemoteStateConsulConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.DataTerraformRemoteStateConsul",
		[]interface{}{scope, id, config},
		d,
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
func DataTerraformRemoteStateConsul_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateDataTerraformRemoteStateConsul_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.DataTerraformRemoteStateConsul",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func DataTerraformRemoteStateConsul_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateDataTerraformRemoteStateConsul_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.DataTerraformRemoteStateConsul",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func DataTerraformRemoteStateConsul_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.DataTerraformRemoteStateConsul",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) AddOverride(path *string, value interface{}) {
	if err := d.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		d,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) Get(output *string) IResolvable {
	if err := d.validateGetParameters(output); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		d,
		"get",
		[]interface{}{output},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) GetBoolean(output *string) IResolvable {
	if err := d.validateGetBooleanParameters(output); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		d,
		"getBoolean",
		[]interface{}{output},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) GetList(output *string) *[]*string {
	if err := d.validateGetListParameters(output); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		d,
		"getList",
		[]interface{}{output},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) GetNumber(output *string) *float64 {
	if err := d.validateGetNumberParameters(output); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		d,
		"getNumber",
		[]interface{}{output},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) GetString(output *string) *string {
	if err := d.validateGetStringParameters(output); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		d,
		"getString",
		[]interface{}{output},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) OverrideLogicalId(newLogicalId *string) {
	if err := d.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		d,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		d,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) ToHclTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		d,
		"toHclTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		d,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		d,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		d,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

