// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Less oft-needed functions to manipulate Tokens.
// Experimental.
type Tokenization interface {
}

// The jsii proxy struct for Tokenization
type jsiiProxy_Tokenization struct {
	_ byte // padding
}

// Experimental.
func NewTokenization() Tokenization {
	_init_.Initialize()

	j := jsiiProxy_Tokenization{}

	_jsii_.Create(
		"cdktf.Tokenization",
		nil, // no parameters
		&j,
	)

	return &j
}

// Experimental.
func NewTokenization_Override(t Tokenization) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.Tokenization",
		nil, // no parameters
		t,
	)
}

// Return whether the given object is an IResolvable object.
//
// This is different from Token.isUnresolved() which will also check for
// encoded Tokens, whereas this method will only do a type check on the given
// object.
// Experimental.
func Tokenization_IsResolvable(obj interface{}) *bool {
	_init_.Initialize()

	if err := validateTokenization_IsResolvableParameters(obj); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.Tokenization",
		"isResolvable",
		[]interface{}{obj},
		&returns,
	)

	return returns
}

// Resolves an object by evaluating all tokens and removing any undefined or empty objects or arrays.
//
// Values can only be primitives, arrays or tokens. Other objects (i.e. with methods) will be rejected.
// Experimental.
func Tokenization_Resolve(obj interface{}, options *ResolveOptions) interface{} {
	_init_.Initialize()

	if err := validateTokenization_ResolveParameters(obj, options); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.Tokenization",
		"resolve",
		[]interface{}{obj, options},
		&returns,
	)

	return returns
}

// Reverse any value into Resolvables, if possible.
// Experimental.
func Tokenization_Reverse(x interface{}) *[]IResolvable {
	_init_.Initialize()

	if err := validateTokenization_ReverseParameters(x); err != nil {
		panic(err)
	}
	var returns *[]IResolvable

	_jsii_.StaticInvoke(
		"cdktf.Tokenization",
		"reverse",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Un-encode a Tokenized value from a list.
// Experimental.
func Tokenization_ReverseList(l *[]*string) IResolvable {
	_init_.Initialize()

	if err := validateTokenization_ReverseListParameters(l); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.Tokenization",
		"reverseList",
		[]interface{}{l},
		&returns,
	)

	return returns
}

// Un-encode a Tokenized value from a map.
// Experimental.
func Tokenization_ReverseMap(m *map[string]interface{}) IResolvable {
	_init_.Initialize()

	if err := validateTokenization_ReverseMapParameters(m); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.Tokenization",
		"reverseMap",
		[]interface{}{m},
		&returns,
	)

	return returns
}

// Un-encode a Tokenized value from a number.
// Experimental.
func Tokenization_ReverseNumber(n *float64) IResolvable {
	_init_.Initialize()

	if err := validateTokenization_ReverseNumberParameters(n); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.Tokenization",
		"reverseNumber",
		[]interface{}{n},
		&returns,
	)

	return returns
}

// Un-encode a Tokenized value from a list.
// Experimental.
func Tokenization_ReverseNumberList(l *[]*float64) IResolvable {
	_init_.Initialize()

	if err := validateTokenization_ReverseNumberListParameters(l); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.Tokenization",
		"reverseNumberList",
		[]interface{}{l},
		&returns,
	)

	return returns
}

// Un-encode a string potentially containing encoded tokens.
// Experimental.
func Tokenization_ReverseString(s *string) TokenizedStringFragments {
	_init_.Initialize()

	if err := validateTokenization_ReverseStringParameters(s); err != nil {
		panic(err)
	}
	var returns TokenizedStringFragments

	_jsii_.StaticInvoke(
		"cdktf.Tokenization",
		"reverseString",
		[]interface{}{s},
		&returns,
	)

	return returns
}

// Stringify a number directly or lazily if it's a Token.
//
// If it is an object (i.e., { Ref: 'SomeLogicalId' }), return it as-is.
// Experimental.
func Tokenization_StringifyNumber(x *float64) *string {
	_init_.Initialize()

	if err := validateTokenization_StringifyNumberParameters(x); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.Tokenization",
		"stringifyNumber",
		[]interface{}{x},
		&returns,
	)

	return returns
}

