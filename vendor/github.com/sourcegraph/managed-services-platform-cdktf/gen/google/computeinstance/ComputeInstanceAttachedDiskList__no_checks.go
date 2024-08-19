//go:build no_runtime_type_checking

package computeinstance

// Building without runtime type checking enabled, so all the below just return nil

func (c *jsiiProxy_ComputeInstanceAttachedDiskList) validateGetParameters(index *float64) error {
	return nil
}

func (c *jsiiProxy_ComputeInstanceAttachedDiskList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceAttachedDiskList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceAttachedDiskList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceAttachedDiskList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceAttachedDiskList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewComputeInstanceAttachedDiskListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

