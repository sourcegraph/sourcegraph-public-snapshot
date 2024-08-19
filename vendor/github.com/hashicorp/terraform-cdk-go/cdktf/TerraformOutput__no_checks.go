// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TerraformOutput) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformOutput) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateTerraformOutput_IsConstructParameters(x interface{}) error {
	return nil
}

func validateTerraformOutput_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateTerraformOutput_IsTerraformOutputParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_TerraformOutput) validateSetPreconditionParameters(val *Precondition) error {
	return nil
}

func (j *jsiiProxy_TerraformOutput) validateSetStaticIdParameters(val *bool) error {
	return nil
}

func (j *jsiiProxy_TerraformOutput) validateSetValueParameters(val interface{}) error {
	return nil
}

func validateNewTerraformOutputParameters(scope constructs.Construct, id *string, config *TerraformOutputConfig) error {
	return nil
}

