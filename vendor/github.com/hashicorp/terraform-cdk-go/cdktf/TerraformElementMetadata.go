// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type TerraformElementMetadata struct {
	// Experimental.
	Path *string `field:"required" json:"path" yaml:"path"`
	// Experimental.
	StackTrace *[]*string `field:"required" json:"stackTrace" yaml:"stackTrace"`
	// Experimental.
	UniqueId *string `field:"required" json:"uniqueId" yaml:"uniqueId"`
}

