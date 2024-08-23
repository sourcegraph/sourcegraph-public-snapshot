// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_StringListList) validateAllWithMapKeyParameters(mapKeyAttributeName *string) error {
	return nil
}

func (s *jsiiProxy_StringListList) validateGetParameters(index *float64) error {
	return nil
}

func (s *jsiiProxy_StringListList) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_StringListList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_StringListList) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_StringListList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewStringListListParameters(terraformResource IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

