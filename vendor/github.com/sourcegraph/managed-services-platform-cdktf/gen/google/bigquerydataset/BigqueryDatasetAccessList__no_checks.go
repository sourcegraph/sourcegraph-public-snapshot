//go:build no_runtime_type_checking

package bigquerydataset

// Building without runtime type checking enabled, so all the below just return nil

func (b *jsiiProxy_BigqueryDatasetAccessList) validateGetParameters(index *float64) error {
	return nil
}

func (b *jsiiProxy_BigqueryDatasetAccessList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_BigqueryDatasetAccessList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_BigqueryDatasetAccessList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_BigqueryDatasetAccessList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_BigqueryDatasetAccessList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewBigqueryDatasetAccessListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

