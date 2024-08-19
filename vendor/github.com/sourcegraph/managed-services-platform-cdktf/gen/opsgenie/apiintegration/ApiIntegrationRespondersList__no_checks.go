//go:build no_runtime_type_checking

package apiintegration

// Building without runtime type checking enabled, so all the below just return nil

func (a *jsiiProxy_ApiIntegrationRespondersList) validateGetParameters(index *float64) error {
	return nil
}

func (a *jsiiProxy_ApiIntegrationRespondersList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_ApiIntegrationRespondersList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_ApiIntegrationRespondersList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_ApiIntegrationRespondersList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_ApiIntegrationRespondersList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewApiIntegrationRespondersListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

