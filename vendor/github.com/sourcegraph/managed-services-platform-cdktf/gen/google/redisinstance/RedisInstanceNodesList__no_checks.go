//go:build no_runtime_type_checking

package redisinstance

// Building without runtime type checking enabled, so all the below just return nil

func (r *jsiiProxy_RedisInstanceNodesList) validateGetParameters(index *float64) error {
	return nil
}

func (r *jsiiProxy_RedisInstanceNodesList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_RedisInstanceNodesList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_RedisInstanceNodesList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_RedisInstanceNodesList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewRedisInstanceNodesListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

