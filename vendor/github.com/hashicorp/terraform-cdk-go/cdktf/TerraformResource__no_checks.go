// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TerraformResource) validateAddMoveTargetParameters(moveTarget *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateGetAnyMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateGetBooleanAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateGetBooleanMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateGetListAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateGetNumberAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateGetNumberListAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateGetNumberMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateGetStringAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateGetStringMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateImportFromParameters(id *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateInterpolationForAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateMoveFromIdParameters(id *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateMoveToParameters(moveTarget *string, index interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateMoveToIdParameters(id *string) error {
	return nil
}

func (t *jsiiProxy_TerraformResource) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateTerraformResource_IsConstructParameters(x interface{}) error {
	return nil
}

func validateTerraformResource_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateTerraformResource_IsTerraformResourceParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_TerraformResource) validateSetConnectionParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_TerraformResource) validateSetCountParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_TerraformResource) validateSetLifecycleParameters(val *TerraformResourceLifecycle) error {
	return nil
}

func (j *jsiiProxy_TerraformResource) validateSetProvisionersParameters(val *[]interface{}) error {
	return nil
}

func validateNewTerraformResourceParameters(scope constructs.Construct, id *string, config *TerraformResourceConfig) error {
	return nil
}

