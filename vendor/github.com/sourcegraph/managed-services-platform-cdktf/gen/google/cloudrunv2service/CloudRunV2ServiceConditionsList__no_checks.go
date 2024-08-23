//go:build no_runtime_type_checking

package cloudrunv2service

// Building without runtime type checking enabled, so all the below just return nil

func (c *jsiiProxy_CloudRunV2ServiceConditionsList) validateGetParameters(index *float64) error {
	return nil
}

func (c *jsiiProxy_CloudRunV2ServiceConditionsList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_CloudRunV2ServiceConditionsList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_CloudRunV2ServiceConditionsList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_CloudRunV2ServiceConditionsList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewCloudRunV2ServiceConditionsListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

