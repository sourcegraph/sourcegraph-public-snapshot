//go:build no_runtime_type_checking

package provider

// Building without runtime type checking enabled, so all the below just return nil

func (n *jsiiProxy_Nobl9Provider) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (n *jsiiProxy_Nobl9Provider) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateNobl9Provider_IsConstructParameters(x interface{}) error {
	return nil
}

func validateNobl9Provider_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNobl9Provider_IsTerraformProviderParameters(x interface{}) error {
	return nil
}

func validateNewNobl9ProviderParameters(scope constructs.Construct, id *string, config *Nobl9ProviderConfig) error {
	return nil
}

