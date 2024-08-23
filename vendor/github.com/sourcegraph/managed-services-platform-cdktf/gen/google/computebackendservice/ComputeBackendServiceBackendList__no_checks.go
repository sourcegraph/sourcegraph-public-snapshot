//go:build no_runtime_type_checking

package computebackendservice

// Building without runtime type checking enabled, so all the below just return nil

func (c *jsiiProxy_ComputeBackendServiceBackendList) validateGetParameters(index *float64) error {
	return nil
}

func (c *jsiiProxy_ComputeBackendServiceBackendList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_ComputeBackendServiceBackendList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_ComputeBackendServiceBackendList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_ComputeBackendServiceBackendList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_ComputeBackendServiceBackendList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewComputeBackendServiceBackendListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

