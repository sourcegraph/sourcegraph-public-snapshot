//go:build no_runtime_type_checking

package sqldatabaseinstance

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_SqlDatabaseInstanceIpAddressList) validateGetParameters(index *float64) error {
	return nil
}

func (s *jsiiProxy_SqlDatabaseInstanceIpAddressList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_SqlDatabaseInstanceIpAddressList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_SqlDatabaseInstanceIpAddressList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_SqlDatabaseInstanceIpAddressList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewSqlDatabaseInstanceIpAddressListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

