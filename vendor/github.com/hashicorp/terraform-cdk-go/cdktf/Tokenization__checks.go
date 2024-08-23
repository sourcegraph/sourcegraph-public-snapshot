// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func validateTokenization_IsResolvableParameters(obj interface{}) error {
	if obj == nil {
		return fmt.Errorf("parameter obj is required, but nil was provided")
	}

	return nil
}

func validateTokenization_ResolveParameters(obj interface{}, options *ResolveOptions) error {
	if obj == nil {
		return fmt.Errorf("parameter obj is required, but nil was provided")
	}

	if options == nil {
		return fmt.Errorf("parameter options is required, but nil was provided")
	}
	if err := _jsii_.ValidateStruct(options, func() string { return "parameter options" }); err != nil {
		return err
	}

	return nil
}

func validateTokenization_ReverseParameters(x interface{}) error {
	if x == nil {
		return fmt.Errorf("parameter x is required, but nil was provided")
	}

	return nil
}

func validateTokenization_ReverseListParameters(l *[]*string) error {
	if l == nil {
		return fmt.Errorf("parameter l is required, but nil was provided")
	}

	return nil
}

func validateTokenization_ReverseMapParameters(m *map[string]interface{}) error {
	if m == nil {
		return fmt.Errorf("parameter m is required, but nil was provided")
	}

	return nil
}

func validateTokenization_ReverseNumberParameters(n *float64) error {
	if n == nil {
		return fmt.Errorf("parameter n is required, but nil was provided")
	}

	return nil
}

func validateTokenization_ReverseNumberListParameters(l *[]*float64) error {
	if l == nil {
		return fmt.Errorf("parameter l is required, but nil was provided")
	}

	return nil
}

func validateTokenization_ReverseStringParameters(s *string) error {
	if s == nil {
		return fmt.Errorf("parameter s is required, but nil was provided")
	}

	return nil
}

func validateTokenization_StringifyNumberParameters(x *float64) error {
	if x == nil {
		return fmt.Errorf("parameter x is required, but nil was provided")
	}

	return nil
}

