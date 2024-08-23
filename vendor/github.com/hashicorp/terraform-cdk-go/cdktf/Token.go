// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Represents a special or lazily-evaluated value.
//
// Can be used to delay evaluation of a certain value in case, for example,
// that it requires some context or late-bound data. Can also be used to
// mark values that need special processing at document rendering time.
//
// Tokens can be embedded into strings while retaining their original
// semantics.
// Experimental.
type Token interface {
}

// The jsii proxy struct for Token
type jsiiProxy_Token struct {
	_ byte // padding
}

// Experimental.
func NewToken() Token {
	_init_.Initialize()

	j := jsiiProxy_Token{}

	_jsii_.Create(
		"cdktf.Token",
		nil, // no parameters
		&j,
	)

	return &j
}

// Experimental.
func NewToken_Override(t Token) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.Token",
		nil, // no parameters
		t,
	)
}

// Return a resolvable representation of the given value.
// Experimental.
func Token_AsAny(value interface{}) IResolvable {
	_init_.Initialize()

	if err := validateToken_AsAnyParameters(value); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.Token",
		"asAny",
		[]interface{}{value},
		&returns,
	)

	return returns
}

// Return a reversible map representation of this token.
// Experimental.
func Token_AsAnyMap(value interface{}, options *EncodingOptions) *map[string]interface{} {
	_init_.Initialize()

	if err := validateToken_AsAnyMapParameters(value, options); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.StaticInvoke(
		"cdktf.Token",
		"asAnyMap",
		[]interface{}{value, options},
		&returns,
	)

	return returns
}

// Return a reversible map representation of this token.
// Experimental.
func Token_AsBooleanMap(value interface{}, options *EncodingOptions) *map[string]*bool {
	_init_.Initialize()

	if err := validateToken_AsBooleanMapParameters(value, options); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.StaticInvoke(
		"cdktf.Token",
		"asBooleanMap",
		[]interface{}{value, options},
		&returns,
	)

	return returns
}

// Return a reversible list representation of this token.
// Experimental.
func Token_AsList(value interface{}, options *EncodingOptions) *[]*string {
	_init_.Initialize()

	if err := validateToken_AsListParameters(value, options); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.Token",
		"asList",
		[]interface{}{value, options},
		&returns,
	)

	return returns
}

// Return a reversible map representation of this token.
// Experimental.
func Token_AsMap(value interface{}, mapValue interface{}, options *EncodingOptions) *map[string]interface{} {
	_init_.Initialize()

	if err := validateToken_AsMapParameters(value, mapValue, options); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.StaticInvoke(
		"cdktf.Token",
		"asMap",
		[]interface{}{value, mapValue, options},
		&returns,
	)

	return returns
}

// Return a reversible number representation of this token.
// Experimental.
func Token_AsNumber(value interface{}) *float64 {
	_init_.Initialize()

	if err := validateToken_AsNumberParameters(value); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.Token",
		"asNumber",
		[]interface{}{value},
		&returns,
	)

	return returns
}

// Return a reversible list representation of this token.
// Experimental.
func Token_AsNumberList(value interface{}) *[]*float64 {
	_init_.Initialize()

	if err := validateToken_AsNumberListParameters(value); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.StaticInvoke(
		"cdktf.Token",
		"asNumberList",
		[]interface{}{value},
		&returns,
	)

	return returns
}

// Return a reversible map representation of this token.
// Experimental.
func Token_AsNumberMap(value interface{}, options *EncodingOptions) *map[string]*float64 {
	_init_.Initialize()

	if err := validateToken_AsNumberMapParameters(value, options); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.StaticInvoke(
		"cdktf.Token",
		"asNumberMap",
		[]interface{}{value, options},
		&returns,
	)

	return returns
}

// Return a reversible string representation of this token.
//
// If the Token is initialized with a literal, the stringified value of the
// literal is returned. Otherwise, a special quoted string representation
// of the Token is returned that can be embedded into other strings.
//
// Strings with quoted Tokens in them can be restored back into
// complex values with the Tokens restored by calling `resolve()`
// on the string.
// Experimental.
func Token_AsString(value interface{}, options *EncodingOptions) *string {
	_init_.Initialize()

	if err := validateToken_AsStringParameters(value, options); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.Token",
		"asString",
		[]interface{}{value, options},
		&returns,
	)

	return returns
}

// Return a reversible map representation of this token.
// Experimental.
func Token_AsStringMap(value interface{}, options *EncodingOptions) *map[string]*string {
	_init_.Initialize()

	if err := validateToken_AsStringMapParameters(value, options); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.StaticInvoke(
		"cdktf.Token",
		"asStringMap",
		[]interface{}{value, options},
		&returns,
	)

	return returns
}

// Returns true if obj represents an unresolved value.
//
// One of these must be true:
//
// - `obj` is an IResolvable
// - `obj` is a string containing at least one encoded `IResolvable`
// - `obj` is either an encoded number or list
//
// This does NOT recurse into lists or objects to see if they
// containing resolvables.
// Experimental.
func Token_IsUnresolved(obj interface{}) *bool {
	_init_.Initialize()

	if err := validateToken_IsUnresolvedParameters(obj); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"cdktf.Token",
		"isUnresolved",
		[]interface{}{obj},
		&returns,
	)

	return returns
}

// Return a Token containing a `null` value.
//
// Note: This is different than `undefined`, `nil`, `None` or similar
// as it will end up in the Terraform config and can be used to explicitly
// not set an attribute (which is sometimes required by Terraform providers).
//
// Returns: a Token resolving to `null` as understood by Terraform.
// Experimental.
func Token_NullValue() IResolvable {
	_init_.Initialize()

	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.Token",
		"nullValue",
		nil, // no parameters
		&returns,
	)

	return returns
}

func Token_ANY_MAP_TOKEN_VALUE() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.Token",
		"ANY_MAP_TOKEN_VALUE",
		&returns,
	)
	return returns
}

func Token_NUMBER_MAP_TOKEN_VALUE() *float64 {
	_init_.Initialize()
	var returns *float64
	_jsii_.StaticGet(
		"cdktf.Token",
		"NUMBER_MAP_TOKEN_VALUE",
		&returns,
	)
	return returns
}

func Token_STRING_MAP_TOKEN_VALUE() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.Token",
		"STRING_MAP_TOKEN_VALUE",
		&returns,
	)
	return returns
}

