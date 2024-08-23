// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (d *jsiiProxy_DataResource) validateAddMoveTargetParameters(moveTarget *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateGetAnyMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateGetBooleanAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateGetBooleanMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateGetListAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateGetNumberAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateGetNumberListAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateGetNumberMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateGetStringAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateGetStringMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateImportFromParameters(id *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateInterpolationForAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateMoveFromIdParameters(id *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateMoveToParameters(moveTarget *string, index interface{}) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateMoveToIdParameters(id *string) error {
	return nil
}

func (d *jsiiProxy_DataResource) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateDataResource_GenerateConfigForImportParameters(scope constructs.Construct, importToId *string, importFromId *string) error {
	return nil
}

func validateDataResource_IsConstructParameters(x interface{}) error {
	return nil
}

func validateDataResource_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateDataResource_IsTerraformResourceParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_DataResource) validateSetConnectionParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_DataResource) validateSetCountParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_DataResource) validateSetInputParameters(val *map[string]interface{}) error {
	return nil
}

func (j *jsiiProxy_DataResource) validateSetLifecycleParameters(val *TerraformResourceLifecycle) error {
	return nil
}

func (j *jsiiProxy_DataResource) validateSetProvisionersParameters(val *[]interface{}) error {
	return nil
}

func (j *jsiiProxy_DataResource) validateSetTriggersReplaceParameters(val *map[string]interface{}) error {
	return nil
}

func validateNewDataResourceParameters(scope constructs.Construct, id *string, config *DataConfig) error {
	return nil
}

