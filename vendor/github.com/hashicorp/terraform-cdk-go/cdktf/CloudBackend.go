// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"

	"github.com/aws/constructs-go/constructs/v10"
)

// The Cloud Backend synthesizes a {@link https://developer.hashicorp.com/terraform/cli/cloud/settings#the-cloud-block cloud block}. The cloud block is a nested block within the top-level terraform settings block. It specifies which Terraform Cloud workspaces to use for the current working directory. The cloud block only affects Terraform CLI's behavior. When Terraform Cloud uses a configuration that contains a cloud block - for example, when a workspace is configured to use a VCS provider directly - it ignores the block and behaves according to its own workspace settings.
// Experimental.
type CloudBackend interface {
	TerraformBackend
	// Experimental.
	CdktfStack() TerraformStack
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	// Experimental.
	Name() *string
	// The tree node.
	// Experimental.
	Node() constructs.Node
	// Experimental.
	RawOverrides() interface{}
	// Experimental.
	AddOverride(path *string, value interface{})
	// Creates a TerraformRemoteState resource that accesses this backend.
	// Experimental.
	GetRemoteStateDataSource(scope constructs.Construct, name *string, _fromStack *string) TerraformRemoteState
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
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for CloudBackend
type jsiiProxy_CloudBackend struct {
	jsiiProxy_TerraformBackend
}

func (j *jsiiProxy_CloudBackend) CdktfStack() TerraformStack {
	var returns TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudBackend) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudBackend) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudBackend) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudBackend) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudBackend) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudBackend) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}


// Experimental.
func NewCloudBackend(scope constructs.Construct, props *CloudBackendConfig) CloudBackend {
	_init_.Initialize()

	if err := validateNewCloudBackendParameters(scope, props); err != nil {
		panic(err)
	}
	j := jsiiProxy_CloudBackend{}

	_jsii_.Create(
		"cdktf.CloudBackend",
		[]interface{}{scope, props},
		&j,
	)

	return &j
}

// Experimental.
func NewCloudBackend_Override(c CloudBackend, scope constructs.Construct, props *CloudBackendConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.CloudBackend",
		[]interface{}{scope, props},
		c,
	)
}

// Experimental.
func CloudBackend_IsBackend(x interface{}) *bool {
	_init_.Initialize()

	if err := validateCloudBackend_IsBackendParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.CloudBackend",
		"isBackend",
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
func CloudBackend_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateCloudBackend_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.CloudBackend",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func CloudBackend_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateCloudBackend_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.CloudBackend",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudBackend) AddOverride(path *string, value interface{}) {
	if err := c.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (c *jsiiProxy_CloudBackend) GetRemoteStateDataSource(scope constructs.Construct, name *string, _fromStack *string) TerraformRemoteState {
	if err := c.validateGetRemoteStateDataSourceParameters(scope, name, _fromStack); err != nil {
		panic(err)
	}
	var returns TerraformRemoteState

	_jsii_.Invoke(
		c,
		"getRemoteStateDataSource",
		[]interface{}{scope, name, _fromStack},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudBackend) OverrideLogicalId(newLogicalId *string) {
	if err := c.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (c *jsiiProxy_CloudBackend) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		c,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudBackend) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudBackend) SynthesizeHclAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"synthesizeHclAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudBackend) ToHclTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toHclTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudBackend) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudBackend) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudBackend) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

