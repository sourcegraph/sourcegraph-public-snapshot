// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TerraformProvider) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformProvider) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateTerraformProvider_IsConstructParameters(x interface{}) error {
	return nil
}

func validateTerraformProvider_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateTerraformProvider_IsTerraformProviderParameters(x interface{}) error {
	return nil
}

func validateNewTerraformProviderParameters(scope constructs.Construct, id *string, config *TerraformProviderConfig) error {
	return nil
}

