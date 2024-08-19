// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TerraformLocal) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformLocal) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateTerraformLocal_IsConstructParameters(x interface{}) error {
	return nil
}

func validateTerraformLocal_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_TerraformLocal) validateSetExpressionParameters(val interface{}) error {
	return nil
}

func validateNewTerraformLocalParameters(scope constructs.Construct, id *string, expression interface{}) error {
	return nil
}

