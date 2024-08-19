// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (d *jsiiProxy_DataTerraformRemoteStateLocal) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateLocal) validateGetParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateLocal) validateGetBooleanParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateLocal) validateGetListParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateLocal) validateGetNumberParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateLocal) validateGetStringParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateLocal) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateDataTerraformRemoteStateLocal_IsConstructParameters(x interface{}) error {
	return nil
}

func validateDataTerraformRemoteStateLocal_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewDataTerraformRemoteStateLocalParameters(scope constructs.Construct, id *string, config *DataTerraformRemoteStateLocalConfig) error {
	return nil
}

