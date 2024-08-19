// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (i *jsiiProxy_ITerraformResource) validateInterpolationForAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (j *jsiiProxy_ITerraformResource) validateSetCountParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_ITerraformResource) validateSetLifecycleParameters(val *TerraformResourceLifecycle) error {
	return nil
}

