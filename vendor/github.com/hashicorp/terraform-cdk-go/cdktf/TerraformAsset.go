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
type TerraformAsset interface {
	constructs.Construct
	// Experimental.
	AssetHash() *string
	// Experimental.
	SetAssetHash(val *string)
	// Name of the asset.
	// Experimental.
	FileName() *string
	// The tree node.
	// Experimental.
	Node() constructs.Node
	// The path relative to the root of the terraform directory in posix format Use this property to reference the asset.
	// Experimental.
	Path() *string
	// Experimental.
	Type() AssetType
	// Experimental.
	SetType(val AssetType)
	// Returns a string representation of this construct.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for TerraformAsset
type jsiiProxy_TerraformAsset struct {
	internal.Type__constructsConstruct
}

func (j *jsiiProxy_TerraformAsset) AssetHash() *string {
	var returns *string
	_jsii_.Get(
		j,
		"assetHash",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformAsset) FileName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fileName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformAsset) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformAsset) Path() *string {
	var returns *string
	_jsii_.Get(
		j,
		"path",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TerraformAsset) Type() AssetType {
	var returns AssetType
	_jsii_.Get(
		j,
		"type",
		&returns,
	)
	return returns
}


// A Terraform Asset takes a file or directory outside of the CDK for Terraform context and moves it into it.
//
// Assets copy referenced files into the stacks context for further usage in other resources.
// Experimental.
func NewTerraformAsset(scope constructs.Construct, id *string, config *TerraformAssetConfig) TerraformAsset {
	_init_.Initialize()

	if err := validateNewTerraformAssetParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_TerraformAsset{}

	_jsii_.Create(
		"cdktf.TerraformAsset",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// A Terraform Asset takes a file or directory outside of the CDK for Terraform context and moves it into it.
//
// Assets copy referenced files into the stacks context for further usage in other resources.
// Experimental.
func NewTerraformAsset_Override(t TerraformAsset, scope constructs.Construct, id *string, config *TerraformAssetConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.TerraformAsset",
		[]interface{}{scope, id, config},
		t,
	)
}

func (j *jsiiProxy_TerraformAsset)SetAssetHash(val *string) {
	if err := j.validateSetAssetHashParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"assetHash",
		val,
	)
}

func (j *jsiiProxy_TerraformAsset)SetType(val AssetType) {
	if err := j.validateSetTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"type",
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
func TerraformAsset_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateTerraformAsset_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.TerraformAsset",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TerraformAsset) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		t,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

