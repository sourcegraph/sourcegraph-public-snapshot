// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Experimental.
type LazyBase interface {
	IResolvable
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	// Experimental.
	AddPostProcessor(postProcessor IPostProcessor)
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(context IResolveContext) interface{}
	// Experimental.
	ResolveLazy(context IResolveContext) interface{}
	// Turn this Token into JSON.
	//
	// Called automatically when JSON.stringify() is called on a Token.
	// Experimental.
	ToJSON() interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for LazyBase
type jsiiProxy_LazyBase struct {
	jsiiProxy_IResolvable
}

func (j *jsiiProxy_LazyBase) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}


// Experimental.
func NewLazyBase_Override(l LazyBase) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.LazyBase",
		nil, // no parameters
		l,
	)
}

func (l *jsiiProxy_LazyBase) AddPostProcessor(postProcessor IPostProcessor) {
	if err := l.validateAddPostProcessorParameters(postProcessor); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		l,
		"addPostProcessor",
		[]interface{}{postProcessor},
	)
}

func (l *jsiiProxy_LazyBase) Resolve(context IResolveContext) interface{} {
	if err := l.validateResolveParameters(context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		l,
		"resolve",
		[]interface{}{context},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LazyBase) ResolveLazy(context IResolveContext) interface{} {
	if err := l.validateResolveLazyParameters(context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		l,
		"resolveLazy",
		[]interface{}{context},
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LazyBase) ToJSON() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		l,
		"toJSON",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (l *jsiiProxy_LazyBase) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		l,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

