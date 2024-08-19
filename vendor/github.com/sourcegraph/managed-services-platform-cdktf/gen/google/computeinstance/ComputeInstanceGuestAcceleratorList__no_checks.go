//go:build no_runtime_type_checking

package computeinstance

// Building without runtime type checking enabled, so all the below just return nil

func (c *jsiiProxy_ComputeInstanceGuestAcceleratorList) validateGetParameters(index *float64) error {
	return nil
}

func (c *jsiiProxy_ComputeInstanceGuestAcceleratorList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceGuestAcceleratorList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceGuestAcceleratorList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceGuestAcceleratorList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceGuestAcceleratorList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewComputeInstanceGuestAcceleratorListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

