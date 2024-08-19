// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TerraformRemoteState) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformRemoteState) validateGetParameters(output *string) error {
	return nil
}

func (t *jsiiProxy_TerraformRemoteState) validateGetBooleanParameters(output *string) error {
	return nil
}

func (t *jsiiProxy_TerraformRemoteState) validateGetListParameters(output *string) error {
	return nil
}

func (t *jsiiProxy_TerraformRemoteState) validateGetNumberParameters(output *string) error {
	return nil
}

func (t *jsiiProxy_TerraformRemoteState) validateGetStringParameters(output *string) error {
	return nil
}

func (t *jsiiProxy_TerraformRemoteState) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateTerraformRemoteState_IsConstructParameters(x interface{}) error {
	return nil
}

func validateTerraformRemoteState_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewTerraformRemoteStateParameters(scope constructs.Construct, id *string, backend *string, config *DataTerraformRemoteStateConfig) error {
	return nil
}

