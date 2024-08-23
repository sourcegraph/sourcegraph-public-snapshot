// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_SwiftBackend) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (s *jsiiProxy_SwiftBackend) validateGetRemoteStateDataSourceParameters(scope constructs.Construct, name *string, _fromStack *string) error {
	return nil
}

func (s *jsiiProxy_SwiftBackend) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateSwiftBackend_IsBackendParameters(x interface{}) error {
	return nil
}

func validateSwiftBackend_IsConstructParameters(x interface{}) error {
	return nil
}

func validateSwiftBackend_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewSwiftBackendParameters(scope constructs.Construct, props *SwiftBackendConfig) error {
	return nil
}

