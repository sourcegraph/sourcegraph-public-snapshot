// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (b *jsiiProxy_BooleanMap) validateLookupParameters(key *string) error {
	return nil
}

func (b *jsiiProxy_BooleanMap) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_BooleanMap) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_BooleanMap) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func validateNewBooleanMapParameters(terraformResource IInterpolatingParent, terraformAttribute *string) error {
	return nil
}

