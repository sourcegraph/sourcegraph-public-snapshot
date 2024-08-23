// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"

	"github.com/aws/constructs-go/constructs/v10"
)

// For migrating past 0.17 where the feature flag for the old id generation logic was removed after being deprecated since 0.15.
// Experimental.
type MigrateIds interface {
	IAspect
	// All aspects can visit an IConstruct.
	// Experimental.
	Visit(node constructs.IConstruct)
}

// The jsii proxy struct for MigrateIds
type jsiiProxy_MigrateIds struct {
	jsiiProxy_IAspect
}

// Experimental.
func NewMigrateIds() MigrateIds {
	_init_.Initialize()

	j := jsiiProxy_MigrateIds{}

	_jsii_.Create(
		"cdktf.MigrateIds",
		nil, // no parameters
		&j,
	)

	return &j
}

// Experimental.
func NewMigrateIds_Override(m MigrateIds) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.MigrateIds",
		nil, // no parameters
		m,
	)
}

func (m *jsiiProxy_MigrateIds) Visit(node constructs.IConstruct) {
	if err := m.validateVisitParameters(node); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"visit",
		[]interface{}{node},
	)
}

