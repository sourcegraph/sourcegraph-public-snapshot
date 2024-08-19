// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TerraformStack) validateAddDependencyParameters(dependency TerraformStack) error {
	return nil
}

func (t *jsiiProxy_TerraformStack) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformStack) validateAllocateLogicalIdParameters(tfElement interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformStack) validateDependsOnParameters(stack TerraformStack) error {
	return nil
}

func (t *jsiiProxy_TerraformStack) validateGetLogicalIdParameters(tfElement interface{}) error {
	return nil
}

func (t *jsiiProxy_TerraformStack) validateRegisterIncomingCrossStackReferenceParameters(fromStack TerraformStack) error {
	return nil
}

func (t *jsiiProxy_TerraformStack) validateRegisterOutgoingCrossStackReferenceParameters(identifier *string) error {
	return nil
}

func validateTerraformStack_IsConstructParameters(x interface{}) error {
	return nil
}

func validateTerraformStack_IsStackParameters(x interface{}) error {
	return nil
}

func validateTerraformStack_OfParameters(construct constructs.IConstruct) error {
	return nil
}

func (j *jsiiProxy_TerraformStack) validateSetDependenciesParameters(val *[]TerraformStack) error {
	return nil
}

func (j *jsiiProxy_TerraformStack) validateSetMoveTargetsParameters(val TerraformResourceTargets) error {
	return nil
}

func (j *jsiiProxy_TerraformStack) validateSetSynthesizerParameters(val IStackSynthesizer) error {
	return nil
}

func validateNewTerraformStackParameters(scope constructs.Construct, id *string) error {
	return nil
}

