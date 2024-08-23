//go:build no_runtime_type_checking

package storagebucket

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_StorageBucketLifecycleRuleList) validateGetParameters(index *float64) error {
	return nil
}

func (s *jsiiProxy_StorageBucketLifecycleRuleList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_StorageBucketLifecycleRuleList) validateSetInternalValueParameters(val interface{}) error {
	return nil
}

func (j *jsiiProxy_StorageBucketLifecycleRuleList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_StorageBucketLifecycleRuleList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_StorageBucketLifecycleRuleList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewStorageBucketLifecycleRuleListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

