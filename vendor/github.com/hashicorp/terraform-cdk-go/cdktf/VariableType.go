// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Experimental.
type VariableType interface {
}

// The jsii proxy struct for VariableType
type jsiiProxy_VariableType struct {
	_ byte // padding
}

// Experimental.
func NewVariableType_Override(v VariableType) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.VariableType",
		nil, // no parameters
		v,
	)
}

// Experimental.
func VariableType_List(type_ *string) *string {
	_init_.Initialize()

	if err := validateVariableType_ListParameters(type_); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.VariableType",
		"list",
		[]interface{}{type_},
		&returns,
	)

	return returns
}

// Experimental.
func VariableType_Map(type_ *string) *string {
	_init_.Initialize()

	if err := validateVariableType_MapParameters(type_); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.VariableType",
		"map",
		[]interface{}{type_},
		&returns,
	)

	return returns
}

// Experimental.
func VariableType_Object(attributes *map[string]*string) *string {
	_init_.Initialize()

	if err := validateVariableType_ObjectParameters(attributes); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.VariableType",
		"object",
		[]interface{}{attributes},
		&returns,
	)

	return returns
}

// Experimental.
func VariableType_Set(type_ *string) *string {
	_init_.Initialize()

	if err := validateVariableType_SetParameters(type_); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.VariableType",
		"set",
		[]interface{}{type_},
		&returns,
	)

	return returns
}

// Experimental.
func VariableType_Tuple(elements ...*string) *string {
	_init_.Initialize()

	args := []interface{}{}
	for _, a := range elements {
		args = append(args, a)
	}

	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.VariableType",
		"tuple",
		args,
		&returns,
	)

	return returns
}

func VariableType_ANY() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"ANY",
		&returns,
	)
	return returns
}

func VariableType_BOOL() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"BOOL",
		&returns,
	)
	return returns
}

func VariableType_LIST() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"LIST",
		&returns,
	)
	return returns
}

func VariableType_LIST_BOOL() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"LIST_BOOL",
		&returns,
	)
	return returns
}

func VariableType_LIST_NUMBER() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"LIST_NUMBER",
		&returns,
	)
	return returns
}

func VariableType_LIST_STRING() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"LIST_STRING",
		&returns,
	)
	return returns
}

func VariableType_MAP() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"MAP",
		&returns,
	)
	return returns
}

func VariableType_MAP_BOOL() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"MAP_BOOL",
		&returns,
	)
	return returns
}

func VariableType_MAP_NUMBER() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"MAP_NUMBER",
		&returns,
	)
	return returns
}

func VariableType_MAP_STRING() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"MAP_STRING",
		&returns,
	)
	return returns
}

func VariableType_NUMBER() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"NUMBER",
		&returns,
	)
	return returns
}

func VariableType_SET() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"SET",
		&returns,
	)
	return returns
}

func VariableType_SET_BOOL() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"SET_BOOL",
		&returns,
	)
	return returns
}

func VariableType_SET_NUMBER() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"SET_NUMBER",
		&returns,
	)
	return returns
}

func VariableType_SET_STRING() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"SET_STRING",
		&returns,
	)
	return returns
}

func VariableType_STRING() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"cdktf.VariableType",
		"STRING",
		&returns,
	)
	return returns
}

