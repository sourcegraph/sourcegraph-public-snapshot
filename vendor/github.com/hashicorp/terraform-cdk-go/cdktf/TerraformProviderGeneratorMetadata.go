// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type TerraformProviderGeneratorMetadata struct {
	// Experimental.
	ProviderName *string `field:"required" json:"providerName" yaml:"providerName"`
	// Experimental.
	ProviderVersion *string `field:"optional" json:"providerVersion" yaml:"providerVersion"`
	// Experimental.
	ProviderVersionConstraint *string `field:"optional" json:"providerVersionConstraint" yaml:"providerVersionConstraint"`
}

