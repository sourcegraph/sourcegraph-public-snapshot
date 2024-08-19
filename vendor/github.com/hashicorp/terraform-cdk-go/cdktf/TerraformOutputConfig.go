// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type TerraformOutputConfig struct {
	// Experimental.
	Value interface{} `field:"required" json:"value" yaml:"value"`
	// Experimental.
	DependsOn *[]ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	Description *string `field:"optional" json:"description" yaml:"description"`
	// Experimental.
	Precondition *Precondition `field:"optional" json:"precondition" yaml:"precondition"`
	// Experimental.
	Sensitive *bool `field:"optional" json:"sensitive" yaml:"sensitive"`
	// If set to true the synthesized Terraform Output will be named after the `id` passed to the constructor instead of the default (TerraformOutput.friendlyUniqueId).
	// Default: false.
	//
	// Experimental.
	StaticId *bool `field:"optional" json:"staticId" yaml:"staticId"`
}

