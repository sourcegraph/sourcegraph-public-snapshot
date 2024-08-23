//go:build no_runtime_type_checking

package clouddeploytarget

// Building without runtime type checking enabled, so all the below just return nil

func (c *jsiiProxy_ClouddeployTargetExecutionConfigsList) validateGetParameters(index *float64) error {
	return nil
}

func (c *jsiiProxy_ClouddeployTargetExecutionConfigsList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_ClouddeployTargetExecutionConfigsList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_ClouddeployTargetExecutionConfigsList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_ClouddeployTargetExecutionConfigsList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_ClouddeployTargetExecutionConfigsList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewClouddeployTargetExecutionConfigsListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

