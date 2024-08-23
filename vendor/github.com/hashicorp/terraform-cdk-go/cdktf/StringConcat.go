// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Converts all fragments to strings and concats those.
//
// Drops 'undefined's.
// Experimental.
type StringConcat interface {
	IFragmentConcatenator
	// Concatenates string fragments.
	// Experimental.
	Join(left interface{}, right interface{}) interface{}
}

// The jsii proxy struct for StringConcat
type jsiiProxy_StringConcat struct {
	jsiiProxy_IFragmentConcatenator
}

// Experimental.
func NewStringConcat() StringConcat {
	_init_.Initialize()

	j := jsiiProxy_StringConcat{}

	_jsii_.Create(
		"cdktf.StringConcat",
		nil, // no parameters
		&j,
	)

	return &j
}

// Experimental.
func NewStringConcat_Override(s StringConcat) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.StringConcat",
		nil, // no parameters
		s,
	)
}

func (s *jsiiProxy_StringConcat) Join(left interface{}, right interface{}) interface{} {
	if err := s.validateJoinParameters(left, right); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		s,
		"join",
		[]interface{}{left, right},
		&returns,
	)

	return returns
}

