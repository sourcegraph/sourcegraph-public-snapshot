//go:build no_runtime_type_checking

package computeinstance

// Building without runtime type checking enabled, so all the below just return nil

func (c *jsiiProxy_ComputeInstanceScratchDiskList) validateGetParameters(index *float64) error {
	return nil
}

func (c *jsiiProxy_ComputeInstanceScratchDiskList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceScratchDiskList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceScratchDiskList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceScratchDiskList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceScratchDiskList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewComputeInstanceScratchDiskListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

