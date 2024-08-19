// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type TerraformResourceMoveById struct {
	// Experimental.
	From *string `field:"required" json:"from" yaml:"from"`
	// Experimental.
	To *string `field:"required" json:"to" yaml:"to"`
}

