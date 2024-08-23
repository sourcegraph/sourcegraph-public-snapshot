// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (a *jsiiProxy_AnyListList) validateAllWithMapKeyParameters(mapKeyAttributeName *string) error {
	return nil
}

func (a *jsiiProxy_AnyListList) validateGetParameters(index *float64) error {
	return nil
}

func (a *jsiiProxy_AnyListList) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_AnyListList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_AnyListList) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_AnyListList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewAnyListListParameters(terraformResource IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

