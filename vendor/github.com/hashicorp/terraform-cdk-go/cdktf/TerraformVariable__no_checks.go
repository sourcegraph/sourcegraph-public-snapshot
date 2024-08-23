// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TerraformVariable) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformVariable) validateAddValidationParameters(validation *TerraformVariableValidationConfig) error {
	return nil
}

func (t *jsiiProxy_TerraformVariable) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateTerraformVariable_IsConstructParameters(x interface{}) error {
	return nil
}

func validateTerraformVariable_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewTerraformVariableParameters(scope constructs.Construct, id *string, config *TerraformVariableConfig) error {
	return nil
}

