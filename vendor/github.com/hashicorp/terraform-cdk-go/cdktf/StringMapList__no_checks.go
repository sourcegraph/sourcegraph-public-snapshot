// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_StringMapList) validateGetParameters(index *float64) error {
	return nil
}

func (s *jsiiProxy_StringMapList) validateInterpolationForAttributeParameters(property *string) error {
	return nil
}

func (s *jsiiProxy_StringMapList) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_StringMapList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_StringMapList) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_StringMapList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewStringMapListParameters(terraformResource IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

