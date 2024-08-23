//go:build no_runtime_type_checking

package provider

// Building without runtime type checking enabled, so all the below just return nil

func (g *jsiiProxy_GoogleBetaProvider) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (g *jsiiProxy_GoogleBetaProvider) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateGoogleBetaProvider_IsConstructParameters(x interface{}) error {
	return nil
}

func validateGoogleBetaProvider_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateGoogleBetaProvider_IsTerraformProviderParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_GoogleBetaProvider) validateSetAddTerraformAttributionLabelParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_GoogleBetaProvider) validateSetBatchingParameters(val *GoogleBetaProviderBatching) error {
	return nil
}

func (j *jsiiProxy_GoogleBetaProvider) validateSetUserProjectOverrideParameters(val interface{}) error {
	return nil
}

func validateNewGoogleBetaProviderParameters(scope constructs.Construct, id *string, config *GoogleBetaProviderConfig) error {
	return nil
}

