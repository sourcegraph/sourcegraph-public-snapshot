// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (b *jsiiProxy_BooleanListMap) validateGetParameters(key *string) error {
	return nil
}

func (b *jsiiProxy_BooleanListMap) validateInterpolationForAttributeParameters(property *string) error {
	return nil
}

func (b *jsiiProxy_BooleanListMap) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_BooleanListMap) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_BooleanListMap) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func validateNewBooleanListMapParameters(terraformResource IInterpolatingParent, terraformAttribute *string) error {
	return nil
}

