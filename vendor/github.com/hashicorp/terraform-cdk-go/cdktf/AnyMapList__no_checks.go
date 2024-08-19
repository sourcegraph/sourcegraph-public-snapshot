// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (a *jsiiProxy_AnyMapList) validateGetParameters(index *float64) error {
	return nil
}

func (a *jsiiProxy_AnyMapList) validateInterpolationForAttributeParameters(property *string) error {
	return nil
}

func (a *jsiiProxy_AnyMapList) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_AnyMapList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_AnyMapList) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_AnyMapList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewAnyMapListParameters(terraformResource IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

