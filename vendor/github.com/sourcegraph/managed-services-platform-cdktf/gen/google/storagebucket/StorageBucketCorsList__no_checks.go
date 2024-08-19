//go:build no_runtime_type_checking

package storagebucket

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_StorageBucketCorsList) validateGetParameters(index *float64) error {
	return nil
}

func (s *jsiiProxy_StorageBucketCorsList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_StorageBucketCorsList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_StorageBucketCorsList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_StorageBucketCorsList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_StorageBucketCorsList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewStorageBucketCorsListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

