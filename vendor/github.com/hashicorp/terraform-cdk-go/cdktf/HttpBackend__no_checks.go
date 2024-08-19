// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (h *jsiiProxy_HttpBackend) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (h *jsiiProxy_HttpBackend) validateGetRemoteStateDataSourceParameters(scope constructs.Construct, name *string, _fromStack *string) error {
	return nil
}

func (h *jsiiProxy_HttpBackend) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateHttpBackend_IsBackendParameters(x interface{}) error {
	return nil
}

func validateHttpBackend_IsConstructParameters(x interface{}) error {
	return nil
}

func validateHttpBackend_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewHttpBackendParameters(scope constructs.Construct, props *HttpBackendConfig) error {
	return nil
}

