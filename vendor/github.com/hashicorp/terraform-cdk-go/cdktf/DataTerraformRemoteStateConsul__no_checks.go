// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (d *jsiiProxy_DataTerraformRemoteStateConsul) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) validateGetParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) validateGetBooleanParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) validateGetListParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) validateGetNumberParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) validateGetStringParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateConsul) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateDataTerraformRemoteStateConsul_IsConstructParameters(x interface{}) error {
	return nil
}

func validateDataTerraformRemoteStateConsul_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewDataTerraformRemoteStateConsulParameters(scope constructs.Construct, id *string, config *DataTerraformRemoteStateConsulConfig) error {
	return nil
}

