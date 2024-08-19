// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf/internal"
)

// A construct which represents a resource.
// Deprecated: - Please use Construct from the constructs package instead.
type Resource interface {
	constructs.Construct
	IResource
	// The tree node.
	// Deprecated: - Please use Construct from the constructs package instead.
	Node() constructs.Node
	// The stack in which this resource is defined.
	// Deprecated: - Please use Construct from the constructs package instead.
	Stack() TerraformStack
	// Returns a string representation of this construct.
	// Deprecated: - Please use Construct from the constructs package instead.
	ToString() *string
}

// The jsii proxy struct for Resource
type jsiiProxy_Resource struct {
	internal.Type__constructsConstruct
	jsiiProxy_IResource
}

func (j *jsiiProxy_Resource) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Resource) Stack() TerraformStack {
	var returns TerraformStack
	_jsii_.Get(
		j,
		"stack",
		&returns,
	)
	return returns
}


// Deprecated: - Please use Construct from the constructs package instead.
func NewResource_Override(r Resource, scope constructs.Construct, id *string) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.Resource",
		[]interface{}{scope, id},
		r,
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
// Deprecated: - Please use Construct from the constructs package instead.
func Resource_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateResource_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.Resource",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Resource) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		r,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

