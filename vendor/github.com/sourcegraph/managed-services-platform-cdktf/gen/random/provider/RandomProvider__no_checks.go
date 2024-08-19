//go:build no_runtime_type_checking

package provider

// Building without runtime type checking enabled, so all the below just return nil

func (r *jsiiProxy_RandomProvider) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (r *jsiiProxy_RandomProvider) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateRandomProvider_IsConstructParameters(x interface{}) error {
	return nil
}

func validateRandomProvider_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateRandomProvider_IsTerraformProviderParameters(x interface{}) error {
	return nil
}

func validateNewRandomProviderParameters(scope constructs.Construct, id *string, config *RandomProviderConfig) error {
	return nil
}

