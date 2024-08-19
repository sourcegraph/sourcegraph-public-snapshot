//go:build no_runtime_type_checking

package computeinstance

// Building without runtime type checking enabled, so all the below just return nil

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceList) validateGetParameters(index *float64) error {
	return nil
}

func (c *jsiiProxy_ComputeInstanceNetworkInterfaceList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_ComputeInstanceNetworkInterfaceList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewComputeInstanceNetworkInterfaceListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

