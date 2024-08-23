// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_StringListMap) validateGetParameters(key *string) error {
	return nil
}

func (s *jsiiProxy_StringListMap) validateInterpolationForAttributeParameters(property *string) error {
	return nil
}

func (s *jsiiProxy_StringListMap) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_StringListMap) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_StringListMap) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func validateNewStringListMapParameters(terraformResource IInterpolatingParent, terraformAttribute *string) error {
	return nil
}

