// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type TerraformVariableConfig struct {
	// Experimental.
	Default interface{} `field:"optional" json:"default" yaml:"default"`
	// Experimental.
	Description *string `field:"optional" json:"description" yaml:"description"`
	// Experimental.
	Nullable *bool `field:"optional" json:"nullable" yaml:"nullable"`
	// Experimental.
	Sensitive *bool `field:"optional" json:"sensitive" yaml:"sensitive"`
	// The type argument in a variable block allows you to restrict the type of value that will be accepted as the value for a variable.
	//
	// If no type constraint is set then a value of any type is accepted.
	//
	// While type constraints are optional, we recommend specifying them; they serve as easy reminders for users of the module, and allow Terraform to return a helpful error message if the wrong type is used.
	//
	// Type constraints are created from a mixture of type keywords and type constructors. The supported type keywords are:
	//
	// - string
	// - number
	// - bool
	//
	// The type constructors allow you to specify complex types such as collections:
	//
	// - list(\<TYPE\>)
	// - set(\<TYPE\>)
	// - map(\<TYPE\>)
	// - object({\<ATTR NAME\> = \<TYPE\>, ... })
	// - tuple([\<TYPE\>, ...])
	//
	// The keyword any may be used to indicate that any type is acceptable. For more information on the meaning and behavior of these different types, as well as detailed information about automatic conversion of complex types, refer to {@link https://developer.hashicorp.com/terraform/language/expressions/type-constraints Type Constraints}.
	//
	// If both the type and default arguments are specified, the given default value must be convertible to the specified type.
	// Experimental.
	Type *string `field:"optional" json:"type" yaml:"type"`
	// Specify arbitrary custom validation rules for a particular variable using a validation block nested within the corresponding variable block.
	// Experimental.
	Validation *[]*TerraformVariableValidationConfig `field:"optional" json:"validation" yaml:"validation"`
}

