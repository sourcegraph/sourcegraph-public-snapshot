// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf/internal"
)

// Experimental.
type IResource interface {
	constructs.IConstruct
	// The stack in which this resource is defined.
	// Experimental.
	Stack() TerraformStack
}

// The jsii proxy for IResource
type jsiiProxy_IResource struct {
	internal.Type__constructsIConstruct
}

func (j *jsiiProxy_IResource) Stack() TerraformStack {
	var returns TerraformStack
	_jsii_.Get(
		j,
		"stack",
		&returns,
	)
	return returns
}

