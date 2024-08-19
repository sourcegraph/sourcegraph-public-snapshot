// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (d *jsiiProxy_DataTerraformRemoteState) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteState) validateGetParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteState) validateGetBooleanParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteState) validateGetListParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteState) validateGetNumberParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteState) validateGetStringParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteState) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateDataTerraformRemoteState_IsConstructParameters(x interface{}) error {
	return nil
}

func validateDataTerraformRemoteState_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewDataTerraformRemoteStateParameters(scope constructs.Construct, id *string, config *DataTerraformRemoteStateRemoteConfig) error {
	return nil
}

