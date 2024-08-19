// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (d *jsiiProxy_DataTerraformRemoteStateGcs) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateGcs) validateGetParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateGcs) validateGetBooleanParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateGcs) validateGetListParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateGcs) validateGetNumberParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateGcs) validateGetStringParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateGcs) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateDataTerraformRemoteStateGcs_IsConstructParameters(x interface{}) error {
	return nil
}

func validateDataTerraformRemoteStateGcs_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewDataTerraformRemoteStateGcsParameters(scope constructs.Construct, id *string, config *DataTerraformRemoteStateGcsConfig) error {
	return nil
}

