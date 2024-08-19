// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func validateToken_AsAnyParameters(value interface{}) error {
	return nil
}

func validateToken_AsAnyMapParameters(value interface{}, options *EncodingOptions) error {
	return nil
}

func validateToken_AsBooleanMapParameters(value interface{}, options *EncodingOptions) error {
	return nil
}

func validateToken_AsListParameters(value interface{}, options *EncodingOptions) error {
	return nil
}

func validateToken_AsMapParameters(value interface{}, mapValue interface{}, options *EncodingOptions) error {
	return nil
}

func validateToken_AsNumberParameters(value interface{}) error {
	return nil
}

func validateToken_AsNumberListParameters(value interface{}) error {
	return nil
}

func validateToken_AsNumberMapParameters(value interface{}, options *EncodingOptions) error {
	return nil
}

func validateToken_AsStringParameters(value interface{}, options *EncodingOptions) error {
	return nil
}

func validateToken_AsStringMapParameters(value interface{}, options *EncodingOptions) error {
	return nil
}

func validateToken_IsUnresolvedParameters(obj interface{}) error {
	return nil
}

