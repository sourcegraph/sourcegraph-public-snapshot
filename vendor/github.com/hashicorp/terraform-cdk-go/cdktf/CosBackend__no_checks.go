// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (c *jsiiProxy_CosBackend) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (c *jsiiProxy_CosBackend) validateGetRemoteStateDataSourceParameters(scope constructs.Construct, name *string, _fromStack *string) error {
	return nil
}

func (c *jsiiProxy_CosBackend) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateCosBackend_IsBackendParameters(x interface{}) error {
	return nil
}

func validateCosBackend_IsConstructParameters(x interface{}) error {
	return nil
}

func validateCosBackend_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewCosBackendParameters(scope constructs.Construct, props *CosBackendConfig) error {
	return nil
}

