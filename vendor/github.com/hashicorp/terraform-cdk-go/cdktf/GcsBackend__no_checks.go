// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (g *jsiiProxy_GcsBackend) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (g *jsiiProxy_GcsBackend) validateGetRemoteStateDataSourceParameters(scope constructs.Construct, name *string, _fromStack *string) error {
	return nil
}

func (g *jsiiProxy_GcsBackend) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateGcsBackend_IsBackendParameters(x interface{}) error {
	return nil
}

func validateGcsBackend_IsConstructParameters(x interface{}) error {
	return nil
}

func validateGcsBackend_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewGcsBackendParameters(scope constructs.Construct, props *GcsBackendConfig) error {
	return nil
}

