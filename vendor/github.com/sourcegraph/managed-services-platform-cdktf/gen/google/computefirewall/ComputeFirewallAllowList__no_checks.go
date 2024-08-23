//go:build no_runtime_type_checking

package computefirewall

// Building without runtime type checking enabled, so all the below just return nil

func (c *jsiiProxy_ComputeFirewallAllowList) validateGetParameters(index *float64) error {
	return nil
}

func (c *jsiiProxy_ComputeFirewallAllowList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_ComputeFirewallAllowList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_ComputeFirewallAllowList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_ComputeFirewallAllowList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_ComputeFirewallAllowList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewComputeFirewallAllowListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

