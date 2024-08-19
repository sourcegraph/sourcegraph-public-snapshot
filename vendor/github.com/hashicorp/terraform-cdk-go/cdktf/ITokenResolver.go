// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

// How to resolve tokens.
// Experimental.
type ITokenResolver interface {
	// Resolve a tokenized list.
	// Experimental.
	ResolveList(l *[]*string, context IResolveContext) interface{}
	// Resolve a tokenized map.
	// Experimental.
	ResolveMap(m *map[string]interface{}, context IResolveContext) interface{}
	// Resolve a tokenized number list.
	// Experimental.
	ResolveNumberList(l *[]*float64, context IResolveContext) interface{}
	// Resolve a string with at least one stringified token in it.
	//
	// (May use concatenation).
	// Experimental.
	ResolveString(s TokenizedStringFragments, context IResolveContext) interface{}
	// Resolve a single token.
	// Experimental.
	ResolveToken(t IResolvable, context IResolveContext, postProcessor IPostProcessor) interface{}
}

// The jsii proxy for ITokenResolver
type jsiiProxy_ITokenResolver struct {
	_ byte // padding
}

func (i *jsiiProxy_ITokenResolver) ResolveList(l *[]*string, context IResolveContext) interface{} {
	if err := i.validateResolveListParameters(l, context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		i,
		"resolveList",
		[]interface{}{l, context},
		&returns,
	)

	return returns
}

func (i *jsiiProxy_ITokenResolver) ResolveMap(m *map[string]interface{}, context IResolveContext) interface{} {
	if err := i.validateResolveMapParameters(m, context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		i,
		"resolveMap",
		[]interface{}{m, context},
		&returns,
	)

	return returns
}

func (i *jsiiProxy_ITokenResolver) ResolveNumberList(l *[]*float64, context IResolveContext) interface{} {
	if err := i.validateResolveNumberListParameters(l, context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		i,
		"resolveNumberList",
		[]interface{}{l, context},
		&returns,
	)

	return returns
}

func (i *jsiiProxy_ITokenResolver) ResolveString(s TokenizedStringFragments, context IResolveContext) interface{} {
	if err := i.validateResolveStringParameters(s, context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		i,
		"resolveString",
		[]interface{}{s, context},
		&returns,
	)

	return returns
}

func (i *jsiiProxy_ITokenResolver) ResolveToken(t IResolvable, context IResolveContext, postProcessor IPostProcessor) interface{} {
	if err := i.validateResolveTokenParameters(t, context, postProcessor); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		i,
		"resolveToken",
		[]interface{}{t, context, postProcessor},
		&returns,
	)

	return returns
}

