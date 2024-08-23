// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Default resolver implementation.
// Experimental.
type DefaultTokenResolver interface {
	ITokenResolver
	// Resolves a list of string.
	// Experimental.
	ResolveList(xs *[]*string, context IResolveContext) interface{}
	// Resolves a map token.
	// Experimental.
	ResolveMap(xs *map[string]interface{}, context IResolveContext) interface{}
	// Resolves a list of numbers.
	// Experimental.
	ResolveNumberList(xs *[]*float64, context IResolveContext) interface{}
	// Resolve string fragments to Tokens.
	// Experimental.
	ResolveString(fragments TokenizedStringFragments, context IResolveContext) interface{}
	// Default Token resolution.
	//
	// Resolve the Token, recurse into whatever it returns,
	// then finally post-process it.
	// Experimental.
	ResolveToken(t IResolvable, context IResolveContext, postProcessor IPostProcessor) interface{}
}

// The jsii proxy struct for DefaultTokenResolver
type jsiiProxy_DefaultTokenResolver struct {
	jsiiProxy_ITokenResolver
}

// Resolves tokens.
// Experimental.
func NewDefaultTokenResolver(concat IFragmentConcatenator) DefaultTokenResolver {
	_init_.Initialize()

	if err := validateNewDefaultTokenResolverParameters(concat); err != nil {
		panic(err)
	}
	j := jsiiProxy_DefaultTokenResolver{}

	_jsii_.Create(
		"cdktf.DefaultTokenResolver",
		[]interface{}{concat},
		&j,
	)

	return &j
}

// Resolves tokens.
// Experimental.
func NewDefaultTokenResolver_Override(d DefaultTokenResolver, concat IFragmentConcatenator) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.DefaultTokenResolver",
		[]interface{}{concat},
		d,
	)
}

func (d *jsiiProxy_DefaultTokenResolver) ResolveList(xs *[]*string, context IResolveContext) interface{} {
	if err := d.validateResolveListParameters(xs, context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		d,
		"resolveList",
		[]interface{}{xs, context},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DefaultTokenResolver) ResolveMap(xs *map[string]interface{}, context IResolveContext) interface{} {
	if err := d.validateResolveMapParameters(xs, context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		d,
		"resolveMap",
		[]interface{}{xs, context},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DefaultTokenResolver) ResolveNumberList(xs *[]*float64, context IResolveContext) interface{} {
	if err := d.validateResolveNumberListParameters(xs, context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		d,
		"resolveNumberList",
		[]interface{}{xs, context},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DefaultTokenResolver) ResolveString(fragments TokenizedStringFragments, context IResolveContext) interface{} {
	if err := d.validateResolveStringParameters(fragments, context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		d,
		"resolveString",
		[]interface{}{fragments, context},
		&returns,
	)

	return returns
}

func (d *jsiiProxy_DefaultTokenResolver) ResolveToken(t IResolvable, context IResolveContext, postProcessor IPostProcessor) interface{} {
	if err := d.validateResolveTokenParameters(t, context, postProcessor); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		d,
		"resolveToken",
		[]interface{}{t, context, postProcessor},
		&returns,
	)

	return returns
}

