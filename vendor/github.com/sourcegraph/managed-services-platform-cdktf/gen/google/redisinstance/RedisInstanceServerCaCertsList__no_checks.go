//go:build no_runtime_type_checking

package redisinstance

// Building without runtime type checking enabled, so all the below just return nil

func (r *jsiiProxy_RedisInstanceServerCaCertsList) validateGetParameters(index *float64) error {
	return nil
}

func (r *jsiiProxy_RedisInstanceServerCaCertsList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_RedisInstanceServerCaCertsList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_RedisInstanceServerCaCertsList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_RedisInstanceServerCaCertsList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewRedisInstanceServerCaCertsListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

