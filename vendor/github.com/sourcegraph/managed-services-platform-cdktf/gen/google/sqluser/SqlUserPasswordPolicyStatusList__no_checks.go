//go:build no_runtime_type_checking

package sqluser

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_SqlUserPasswordPolicyStatusList) validateGetParameters(index *float64) error {
	return nil
}

func (s *jsiiProxy_SqlUserPasswordPolicyStatusList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_SqlUserPasswordPolicyStatusList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_SqlUserPasswordPolicyStatusList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_SqlUserPasswordPolicyStatusList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewSqlUserPasswordPolicyStatusListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

