// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type TerraformHclModuleConfig struct {
	// Experimental.
	DependsOn *[]ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	ForEach ITerraformIterator `field:"optional" json:"forEach" yaml:"forEach"`
	// Experimental.
	Providers *[]interface{} `field:"optional" json:"providers" yaml:"providers"`
	// Experimental.
	SkipAssetCreationFromLocalModules *bool `field:"optional" json:"skipAssetCreationFromLocalModules" yaml:"skipAssetCreationFromLocalModules"`
	// Experimental.
	Source *string `field:"required" json:"source" yaml:"source"`
	// Experimental.
	Version *string `field:"optional" json:"version" yaml:"version"`
	// Experimental.
	Variables *map[string]interface{} `field:"optional" json:"variables" yaml:"variables"`
}

