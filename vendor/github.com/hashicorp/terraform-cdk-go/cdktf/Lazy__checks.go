// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func validateLazy_AnyValueParameters(producer IAnyProducer, options *LazyAnyValueOptions) error {
	if producer == nil {
		return fmt.Errorf("parameter producer is required, but nil was provided")
	}

	if err := _jsii_.ValidateStruct(options, func() string { return "parameter options" }); err != nil {
		return err
	}

	return nil
}

func validateLazy_ListValueParameters(producer IListProducer, options *LazyListValueOptions) error {
	if producer == nil {
		return fmt.Errorf("parameter producer is required, but nil was provided")
	}

	if err := _jsii_.ValidateStruct(options, func() string { return "parameter options" }); err != nil {
		return err
	}

	return nil
}

func validateLazy_NumberValueParameters(producer INumberProducer) error {
	if producer == nil {
		return fmt.Errorf("parameter producer is required, but nil was provided")
	}

	return nil
}

func validateLazy_StringValueParameters(producer IStringProducer, options *LazyStringValueOptions) error {
	if producer == nil {
		return fmt.Errorf("parameter producer is required, but nil was provided")
	}

	if err := _jsii_.ValidateStruct(options, func() string { return "parameter options" }); err != nil {
		return err
	}

	return nil
}

