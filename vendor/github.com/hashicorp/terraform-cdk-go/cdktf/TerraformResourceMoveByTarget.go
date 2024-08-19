// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type TerraformResourceMoveByTarget struct {
	// Experimental.
	MoveTarget *string `field:"required" json:"moveTarget" yaml:"moveTarget"`
	// Experimental.
	Index interface{} `field:"optional" json:"index" yaml:"index"`
}

