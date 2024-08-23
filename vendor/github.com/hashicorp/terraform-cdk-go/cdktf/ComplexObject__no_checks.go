// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (c *jsiiProxy_ComplexObject) validateGetAnyMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (c *jsiiProxy_ComplexObject) validateGetBooleanAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (c *jsiiProxy_ComplexObject) validateGetBooleanMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (c *jsiiProxy_ComplexObject) validateGetListAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (c *jsiiProxy_ComplexObject) validateGetNumberAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (c *jsiiProxy_ComplexObject) validateGetNumberListAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (c *jsiiProxy_ComplexObject) validateGetNumberMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (c *jsiiProxy_ComplexObject) validateGetStringAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (c *jsiiProxy_ComplexObject) validateGetStringMapAttributeParameters(terraformAttribute *string) error {
	return nil
}

func (c *jsiiProxy_ComplexObject) validateInterpolationForAttributeParameters(property *string) error {
	return nil
}

func (c *jsiiProxy_ComplexObject) validateResolveParameters(_context IResolveContext) error {
	return nil
}

func (j *jsiiProxy_ComplexObject) validateSetComplexObjectIndexParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_ComplexObject) validateSetComplexObjectIsFromSetParameters(val *bool) error {
	return nil
}

func (j *jsiiProxy_ComplexObject) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_ComplexObject) validateSetTerraformResourceParameters(val IInterpolatingParent) error {
	return nil
}

func validateNewComplexObjectParameters(terraformResource IInterpolatingParent, terraformAttribute *string, complexObjectIsFromSet *bool, complexObjectIndex interface{}) error {
	return nil
}

