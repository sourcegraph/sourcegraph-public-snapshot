// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (n *jsiiProxy_NumberListList) validateAllWithMapKeyParameters(mapKeyAttributeName *string) error {
	return nil
}

func (n *jsiiProxy_NumberListList) validateGetParameters(index *float64) error {
	return nil
}

func (n *jsiiProxy_NumberListList) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_NumberListList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_NumberListList) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_NumberListList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewNumberListListParameters(terraformResource IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

