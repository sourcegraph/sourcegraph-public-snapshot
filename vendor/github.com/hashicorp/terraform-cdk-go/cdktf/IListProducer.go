// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

// Interface for lazy list producers.
// Experimental.
type IListProducer interface {
	// Produce the list value.
	// Experimental.
	Produce(context IResolveContext) *[]*string
}

// The jsii proxy for IListProducer
type jsiiProxy_IListProducer struct {
	_ byte // padding
}

func (i *jsiiProxy_IListProducer) Produce(context IResolveContext) *[]*string {
	if err := i.validateProduceParameters(context); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		i,
		"produce",
		[]interface{}{context},
		&returns,
	)

	return returns
}

