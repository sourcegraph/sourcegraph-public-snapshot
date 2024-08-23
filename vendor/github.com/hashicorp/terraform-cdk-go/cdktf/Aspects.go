// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"

	"github.com/aws/constructs-go/constructs/v10"
)

// Aspects can be applied to CDK tree scopes and can operate on the tree before synthesis.
// Experimental.
type Aspects interface {
	// The list of aspects which were directly applied on this scope.
	// Experimental.
	All() *[]IAspect
	// Adds an aspect to apply this scope before synthesis.
	// Experimental.
	Add(aspect IAspect)
}

// The jsii proxy struct for Aspects
type jsiiProxy_Aspects struct {
	_ byte // padding
}

func (j *jsiiProxy_Aspects) All() *[]IAspect {
	var returns *[]IAspect
	_jsii_.Get(
		j,
		"all",
		&returns,
	)
	return returns
}


// Returns the `Aspects` object associated with a construct scope.
// Experimental.
func Aspects_Of(scope constructs.IConstruct) Aspects {
	_init_.Initialize()

	if err := validateAspects_OfParameters(scope); err != nil {
		panic(err)
	}
	var returns Aspects

	_jsii_.StaticInvoke(
		"cdktf.Aspects",
		"of",
		[]interface{}{scope},
		&returns,
	)

	return returns
}

func (a *jsiiProxy_Aspects) Add(aspect IAspect) {
	if err := a.validateAddParameters(aspect); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		a,
		"add",
		[]interface{}{aspect},
	)
}

