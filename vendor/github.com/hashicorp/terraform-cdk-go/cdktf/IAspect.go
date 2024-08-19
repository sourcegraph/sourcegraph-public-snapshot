// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"

	"github.com/aws/constructs-go/constructs/v10"
)

// Represents an Aspect.
// Experimental.
type IAspect interface {
	// All aspects can visit an IConstruct.
	// Experimental.
	Visit(node constructs.IConstruct)
}

// The jsii proxy for IAspect
type jsiiProxy_IAspect struct {
	_ byte // padding
}

func (i *jsiiProxy_IAspect) Visit(node constructs.IConstruct) {
	if err := i.validateVisitParameters(node); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		i,
		"visit",
		[]interface{}{node},
	)
}

