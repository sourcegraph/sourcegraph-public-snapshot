// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (s *jsiiProxy_S3Backend) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (s *jsiiProxy_S3Backend) validateGetRemoteStateDataSourceParameters(scope constructs.Construct, name *string, _fromStack *string) error {
	return nil
}

func (s *jsiiProxy_S3Backend) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateS3Backend_IsBackendParameters(x interface{}) error {
	return nil
}

func validateS3Backend_IsConstructParameters(x interface{}) error {
	return nil
}

func validateS3Backend_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewS3BackendParameters(scope constructs.Construct, props *S3BackendConfig) error {
	return nil
}

