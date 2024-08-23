// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type TerraformModuleProvider struct {
	// Experimental.
	ModuleAlias *string `field:"required" json:"moduleAlias" yaml:"moduleAlias"`
	// Experimental.
	Provider TerraformProvider `field:"required" json:"provider" yaml:"provider"`
}

