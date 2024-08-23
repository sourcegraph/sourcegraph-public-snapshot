// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (n *jsiiProxy_NumberListMap) validateGetParameters(key *string) error {
	return nil
}

func (n *jsiiProxy_NumberListMap) validateInterpolationForAttributeParameters(property *string) error {
	return nil
}

func (n *jsiiProxy_NumberListMap) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_NumberListMap) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_NumberListMap) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func validateNewNumberListMapParameters(terraformResource IInterpolatingParent, terraformAttribute *string) error {
	return nil
}

