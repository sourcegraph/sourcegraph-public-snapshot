// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type DataConfig struct {
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// Experimental.
	Count interface{} `field:"optional" json:"count" yaml:"count"`
	// Experimental.
	DependsOn *[]ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	ForEach ITerraformIterator `field:"optional" json:"forEach" yaml:"forEach"`
	// Experimental.
	Lifecycle *TerraformResourceLifecycle `field:"optional" json:"lifecycle" yaml:"lifecycle"`
	// Experimental.
	Provider TerraformProvider `field:"optional" json:"provider" yaml:"provider"`
	// Experimental.
	Provisioners *[]interface{} `field:"optional" json:"provisioners" yaml:"provisioners"`
	// (Optional) A value which will be stored in the instance state, and reflected in the output attribute after apply.
	//
	// https://developer.hashicorp.com/terraform/language/resources/terraform-data#input
	// Experimental.
	Input *map[string]interface{} `field:"optional" json:"input" yaml:"input"`
	// (Optional) A value which is stored in the instance state, and will force replacement when the value changes.
	//
	// https://developer.hashicorp.com/terraform/language/resources/terraform-data#triggers_replace
	// Experimental.
	TriggersReplace *map[string]interface{} `field:"optional" json:"triggersReplace" yaml:"triggersReplace"`
}

