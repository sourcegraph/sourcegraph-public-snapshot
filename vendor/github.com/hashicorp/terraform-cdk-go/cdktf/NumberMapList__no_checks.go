// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (n *jsiiProxy_NumberMapList) validateGetParameters(index *float64) error {
	return nil
}

func (n *jsiiProxy_NumberMapList) validateInterpolationForAttributeParameters(property *string) error {
	return nil
}

func (n *jsiiProxy_NumberMapList) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_NumberMapList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_NumberMapList) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_NumberMapList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewNumberMapListParameters(terraformResource IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

