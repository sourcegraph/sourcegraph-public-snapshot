// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"

	"github.com/aws/constructs-go/constructs/v10"
)

// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
type SwiftBackend interface {
	TerraformBackend
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	CdktfStack() TerraformStack
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ConstructNodeMetadata() *map[string]interface{}
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	Fqn() *string
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	FriendlyUniqueId() *string
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	Name() *string
	// The tree node.
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	Node() constructs.Node
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	RawOverrides() interface{}
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	AddOverride(path *string, value interface{})
	// Creates a TerraformRemoteState resource that accesses this backend.
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	GetRemoteStateDataSource(scope constructs.Construct, name *string, _fromStack *string) TerraformRemoteState
	// Overrides the auto-generated logical ID with a specific ID.
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	OverrideLogicalId(newLogicalId *string)
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ResetOverrideLogicalId()
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	SynthesizeAttributes() *map[string]interface{}
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	SynthesizeHclAttributes() *map[string]interface{}
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ToHclTerraform() interface{}
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
	ToTerraform() interface{}
}

// The jsii proxy struct for SwiftBackend
type jsiiProxy_SwiftBackend struct {
	jsiiProxy_TerraformBackend
}

func (j *jsiiProxy_SwiftBackend) CdktfStack() TerraformStack {
	var returns TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SwiftBackend) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SwiftBackend) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SwiftBackend) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SwiftBackend) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SwiftBackend) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SwiftBackend) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}


// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
func NewSwiftBackend(scope constructs.Construct, props *SwiftBackendConfig) SwiftBackend {
	_init_.Initialize()

	if err := validateNewSwiftBackendParameters(scope, props); err != nil {
		panic(err)
	}
	j := jsiiProxy_SwiftBackend{}

	_jsii_.Create(
		"cdktf.SwiftBackend",
		[]interface{}{scope, props},
		&j,
	)

	return &j
}

// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
func NewSwiftBackend_Override(s SwiftBackend, scope constructs.Construct, props *SwiftBackendConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.SwiftBackend",
		[]interface{}{scope, props},
		s,
	)
}

// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
func SwiftBackend_IsBackend(x interface{}) *bool {
	_init_.Initialize()

	if err := validateSwiftBackend_IsBackendParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.SwiftBackend",
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
// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
func SwiftBackend_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateSwiftBackend_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.SwiftBackend",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Deprecated: CDK for Terraform no longer supports the swift backend. Terraform deprecated swift in v1.2.3 and removed it in v1.3.
func SwiftBackend_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateSwiftBackend_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.SwiftBackend",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SwiftBackend) AddOverride(path *string, value interface{}) {
	if err := s.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (s *jsiiProxy_SwiftBackend) GetRemoteStateDataSource(scope constructs.Construct, name *string, _fromStack *string) TerraformRemoteState {
	if err := s.validateGetRemoteStateDataSourceParameters(scope, name, _fromStack); err != nil {
		panic(err)
	}
	var returns TerraformRemoteState

	_jsii_.Invoke(
		s,
		"getRemoteStateDataSource",
		[]interface{}{scope, name, _fromStack},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SwiftBackend) OverrideLogicalId(newLogicalId *string) {
	if err := s.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (s *jsiiProxy_SwiftBackend) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		s,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SwiftBackend) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		s,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SwiftBackend) SynthesizeHclAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		s,
		"synthesizeHclAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SwiftBackend) ToHclTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		s,
		"toHclTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SwiftBackend) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		s,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SwiftBackend) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SwiftBackend) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		s,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

