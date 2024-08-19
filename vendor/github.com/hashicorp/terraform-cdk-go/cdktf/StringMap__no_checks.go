// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_StringMap) validateLookupParameters(key *string) error {
	return nil
}

func (s *jsiiProxy_StringMap) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_StringMap) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_StringMap) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func validateNewStringMapParameters(terraformResource IInterpolatingParent, terraformAttribute *string) error {
	return nil
}

