//go:build no_runtime_type_checking

package provider

// Building without runtime type checking enabled, so all the below just return nil

func (p *jsiiProxy_PostgresqlProvider) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (p *jsiiProxy_PostgresqlProvider) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validatePostgresqlProvider_GenerateConfigForImportParameters(scope constructs.Construct, importToId *string, importFromId *string) error {
	return nil
}

func validatePostgresqlProvider_IsConstructParameters(x interface{}) error {
	return nil
}

func validatePostgresqlProvider_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validatePostgresqlProvider_IsTerraformProviderParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_PostgresqlProvider) validateSetAwsRdsIamAuthParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_PostgresqlProvider) validateSetAzureIdentityAuthParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_PostgresqlProvider) validateSetClientcertParameters(val *PostgresqlProviderClientcert) error {
	return nil
}

func (j *jsiiProxy_PostgresqlProvider) validateSetSuperuserParameters(val interface{}) error {
	return nil
}

func validateNewPostgresqlProviderParameters(scope constructs.Construct, id *string, config *PostgresqlProviderConfig) error {
	return nil
}

