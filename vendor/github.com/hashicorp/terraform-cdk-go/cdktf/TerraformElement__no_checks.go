// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TerraformElement) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformElement) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateTerraformElement_IsConstructParameters(x interface{}) error {
	return nil
}

func validateTerraformElement_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewTerraformElementParameters(scope constructs.Construct, id *string) error {
	return nil
}

