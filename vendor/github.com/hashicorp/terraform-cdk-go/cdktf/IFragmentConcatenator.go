// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

// Function used to concatenate symbols in the target document language.
//
// Interface so it could potentially be exposed over jsii.
// Experimental.
type IFragmentConcatenator interface {
	// Join the fragment on the left and on the right.
	// Experimental.
	Join(left interface{}, right interface{}) interface{}
}

// The jsii proxy for IFragmentConcatenator
type jsiiProxy_IFragmentConcatenator struct {
	_ byte // padding
}

func (i *jsiiProxy_IFragmentConcatenator) Join(left interface{}, right interface{}) interface{} {
	if err := i.validateJoinParameters(left, right); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		i,
		"join",
		[]interface{}{left, right},
		&returns,
	)

	return returns
}

