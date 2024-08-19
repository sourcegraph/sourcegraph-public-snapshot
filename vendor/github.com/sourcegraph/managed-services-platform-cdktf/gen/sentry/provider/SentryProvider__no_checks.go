//go:build no_runtime_type_checking

package provider

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_SentryProvider) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (s *jsiiProxy_SentryProvider) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateSentryProvider_IsConstructParameters(x interface{}) error {
	return nil
}

func validateSentryProvider_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateSentryProvider_IsTerraformProviderParameters(x interface{}) error {
	return nil
}

func validateNewSentryProviderParameters(scope constructs.Construct, id *string, config *SentryProviderConfig) error {
	return nil
}

