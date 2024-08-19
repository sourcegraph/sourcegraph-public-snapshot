// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (c *jsiiProxy_ConsulBackend) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (c *jsiiProxy_ConsulBackend) validateGetRemoteStateDataSourceParameters(scope constructs.Construct, name *string, _fromStack *string) error {
	return nil
}

func (c *jsiiProxy_ConsulBackend) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateConsulBackend_IsBackendParameters(x interface{}) error {
	return nil
}

func validateConsulBackend_IsConstructParameters(x interface{}) error {
	return nil
}

func validateConsulBackend_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewConsulBackendParameters(scope constructs.Construct, props *ConsulBackendConfig) error {
	return nil
}

