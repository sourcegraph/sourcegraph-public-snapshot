//go:build no_runtime_type_checking

package sqldatabaseinstance

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_SqlDatabaseInstanceSettingsDatabaseFlagsList) validateGetParameters(index *float64) error {
	return nil
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsDatabaseFlagsList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsDatabaseFlagsList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsDatabaseFlagsList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsDatabaseFlagsList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsDatabaseFlagsList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewSqlDatabaseInstanceSettingsDatabaseFlagsListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

