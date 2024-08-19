//go:build no_runtime_type_checking

package secretmanagersecret

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_SecretManagerSecretTopicsList) validateGetParameters(index *float64) error {
	return nil
}

func (s *jsiiProxy_SecretManagerSecretTopicsList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_SecretManagerSecretTopicsList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_SecretManagerSecretTopicsList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_SecretManagerSecretTopicsList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_SecretManagerSecretTopicsList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewSecretManagerSecretTopicsListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

