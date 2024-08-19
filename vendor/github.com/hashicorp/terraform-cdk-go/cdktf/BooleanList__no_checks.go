// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (b *jsiiProxy_BooleanList) validateAllWithMapKeyParameters(mapKeyAttributeName *string) error {
	return nil
}

func (b *jsiiProxy_BooleanList) validateGetParameters(index *float64) error {
	return nil
}

func (b *jsiiProxy_BooleanList) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_BooleanList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_BooleanList) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_BooleanList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewBooleanListParameters(terraformResource IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

