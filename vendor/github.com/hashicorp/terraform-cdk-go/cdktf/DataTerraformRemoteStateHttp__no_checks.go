// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (d *jsiiProxy_DataTerraformRemoteStateHttp) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateHttp) validateGetParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateHttp) validateGetBooleanParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateHttp) validateGetListParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateHttp) validateGetNumberParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateHttp) validateGetStringParameters(output *string) error {
	return nil
}

func (d *jsiiProxy_DataTerraformRemoteStateHttp) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateDataTerraformRemoteStateHttp_IsConstructParameters(x interface{}) error {
	return nil
}

func validateDataTerraformRemoteStateHttp_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewDataTerraformRemoteStateHttpParameters(scope constructs.Construct, id *string, config *DataTerraformRemoteStateHttpConfig) error {
	return nil
}

