// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (m *jsiiProxy_MapList) validateInterpolationForAttributeParameters(property *string) error {
	return nil
}

func (m *jsiiProxy_MapList) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_MapList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_MapList) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_MapList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewMapListParameters(terraformResource IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

