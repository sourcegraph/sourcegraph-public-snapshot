//go:build no_runtime_type_checking

package provider

// Building without runtime type checking enabled, so all the below just return nil

func (o *jsiiProxy_OpsgenieProvider) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (o *jsiiProxy_OpsgenieProvider) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateOpsgenieProvider_IsConstructParameters(x interface{}) error {
	return nil
}

func validateOpsgenieProvider_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateOpsgenieProvider_IsTerraformProviderParameters(x interface{}) error {
	return nil
}

func validateNewOpsgenieProviderParameters(scope constructs.Construct, id *string, config *OpsgenieProviderConfig) error {
	return nil
}

