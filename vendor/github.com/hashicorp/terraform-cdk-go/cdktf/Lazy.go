// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Lazily produce a value.
//
// Can be used to return a string, list or numeric value whose actual value
// will only be calculated later, during synthesis.
// Experimental.
type Lazy interface {
}

// The jsii proxy struct for Lazy
type jsiiProxy_Lazy struct {
	_ byte // padding
}

// Experimental.
func NewLazy() Lazy {
	_init_.Initialize()

	j := jsiiProxy_Lazy{}

	_jsii_.Create(
		"cdktf.Lazy",
		nil, // no parameters
		&j,
	)

	return &j
}

// Experimental.
func NewLazy_Override(l Lazy) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.Lazy",
		nil, // no parameters
		l,
	)
}

// Produces a lazy token from an untyped value.
// Experimental.
func Lazy_AnyValue(producer IAnyProducer, options *LazyAnyValueOptions) IResolvable {
	_init_.Initialize()

	if err := validateLazy_AnyValueParameters(producer, options); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.Lazy",
		"anyValue",
		[]interface{}{producer, options},
		&returns,
	)

	return returns
}

// Returns a list-ified token for a lazy value.
// Experimental.
func Lazy_ListValue(producer IListProducer, options *LazyListValueOptions) *[]*string {
	_init_.Initialize()

	if err := validateLazy_ListValueParameters(producer, options); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.Lazy",
		"listValue",
		[]interface{}{producer, options},
		&returns,
	)

	return returns
}

// Returns a numberified token for a lazy value.
// Experimental.
func Lazy_NumberValue(producer INumberProducer) *float64 {
	_init_.Initialize()

	if err := validateLazy_NumberValueParameters(producer); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.Lazy",
		"numberValue",
		[]interface{}{producer},
		&returns,
	)

	return returns
}

// Returns a stringified token for a lazy value.
// Experimental.
func Lazy_StringValue(producer IStringProducer, options *LazyStringValueOptions) *string {
	_init_.Initialize()

	if err := validateLazy_StringValueParameters(producer, options); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.Lazy",
		"stringValue",
		[]interface{}{producer, options},
		&returns,
	)

	return returns
}

