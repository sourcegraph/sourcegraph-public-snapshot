// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (c *jsiiProxy_ComplexList) validateAllWithMapKeyParameters(mapKeyAttributeName *string) error {
	return nil
}

func (c *jsiiProxy_ComplexList) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_ComplexList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_ComplexList) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_ComplexList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewComplexListParameters(terraformResource IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

