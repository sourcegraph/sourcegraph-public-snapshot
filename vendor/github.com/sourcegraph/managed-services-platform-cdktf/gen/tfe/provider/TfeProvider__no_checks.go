//go:build no_runtime_type_checking

package provider

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TfeProvider) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (t *jsiiProxy_TfeProvider) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateTfeProvider_IsConstructParameters(x interface{}) error {
	return nil
}

func validateTfeProvider_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateTfeProvider_IsTerraformProviderParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_TfeProvider) validateSetSslSkipVerifyParameters(val interface{}) error {
	return nil
}

func validateNewTfeProviderParameters(scope constructs.Construct, id *string, config *TfeProviderConfig) error {
	return nil
}

