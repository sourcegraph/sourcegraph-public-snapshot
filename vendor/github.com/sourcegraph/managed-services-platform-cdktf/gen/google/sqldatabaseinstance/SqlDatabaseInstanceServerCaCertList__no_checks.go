//go:build no_runtime_type_checking

package sqldatabaseinstance

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_SqlDatabaseInstanceServerCaCertList) validateGetParameters(index *float64) error {
	return nil
}

func (s *jsiiProxy_SqlDatabaseInstanceServerCaCertList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_SqlDatabaseInstanceServerCaCertList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_SqlDatabaseInstanceServerCaCertList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_SqlDatabaseInstanceServerCaCertList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewSqlDatabaseInstanceServerCaCertListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

