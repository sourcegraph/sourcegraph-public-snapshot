// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TerraformDataSource) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformDataSource) validateGetAnyMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformDataSource) validateGetBooleanAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformDataSource) validateGetBooleanMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformDataSource) validateGetListAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformDataSource) validateGetNumberAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformDataSource) validateGetNumberListAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformDataSource) validateGetNumberMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformDataSource) validateGetStringAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformDataSource) validateGetStringMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformDataSource) validateInterpolationForAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformDataSource) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateTerraformDataSource_IsConstructParameters(x interface{}) error {
	return nil
}

func validateTerraformDataSource_IsTerraformDataSourceParameters(x interface{}) error {
	return nil
}

func validateTerraformDataSource_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_TerraformDataSource) validateSetCountParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_TerraformDataSource) validateSetLifecycleParameters(val *TerraformResourceLifecycle) error {
	return nil
}

func validateNewTerraformDataSourceParameters(scope constructs.Construct, id *string, config *TerraformResourceConfig) error {
	return nil
}

