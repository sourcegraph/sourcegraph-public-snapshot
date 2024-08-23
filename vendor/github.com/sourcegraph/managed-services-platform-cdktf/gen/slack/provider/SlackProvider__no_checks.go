//go:build no_runtime_type_checking

package provider

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_SlackProvider) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (s *jsiiProxy_SlackProvider) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateSlackProvider_IsConstructParameters(x interface{}) error {
	return nil
}

func validateSlackProvider_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateSlackProvider_IsTerraformProviderParameters(x interface{}) error {
	return nil
}

func validateNewSlackProviderParameters(scope constructs.Construct, id *string, config *SlackProviderConfig) error {
	return nil
}

