// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (a *jsiiProxy_AnyListMap) validateGetParameters(key *string) error {
	return nil
}

func (a *jsiiProxy_AnyListMap) validateInterpolationForAttributeParameters(property *string) error {
	return nil
}

func (a *jsiiProxy_AnyListMap) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_AnyListMap) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_AnyListMap) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func validateNewAnyListMapParameters(terraformResource IInterpolatingParent, terraformAttribute *string) error {
	return nil
}

