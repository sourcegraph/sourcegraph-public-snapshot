// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type TerraformResourceLifecycle struct {
	// Experimental.
	CreateBeforeDestroy *bool `field:"optional" json:"createBeforeDestroy" yaml:"createBeforeDestroy"`
	// Experimental.
	IgnoreChanges interface{} `field:"optional" json:"ignoreChanges" yaml:"ignoreChanges"`
	// Experimental.
	Postcondition *[]*Postcondition `field:"optional" json:"postcondition" yaml:"postcondition"`
	// Experimental.
	Precondition *[]*Precondition `field:"optional" json:"precondition" yaml:"precondition"`
	// Experimental.
	PreventDestroy *bool `field:"optional" json:"preventDestroy" yaml:"preventDestroy"`
	// Experimental.
	ReplaceTriggeredBy *[]interface{} `field:"optional" json:"replaceTriggeredBy" yaml:"replaceTriggeredBy"`
}

