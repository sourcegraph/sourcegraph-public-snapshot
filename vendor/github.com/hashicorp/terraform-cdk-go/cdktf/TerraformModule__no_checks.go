// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TerraformModule) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformModule) validateAddProviderParameters(provider interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformModule) validateGetStringParameters(output *string) error {
	return nil
}

func (t *jsiiProxy_TerraformModule) validateInterpolationForOutputParameters(moduleOutput *string) error {
	return nil
}

func (t *jsiiProxy_TerraformModule) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateTerraformModule_IsConstructParameters(x interface{}) error {
	return nil
}

func validateTerraformModule_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewTerraformModuleParameters(scope constructs.Construct, id *string, options *TerraformModuleConfig) error {
	return nil
}

