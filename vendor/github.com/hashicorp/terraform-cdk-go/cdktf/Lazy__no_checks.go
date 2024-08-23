// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func validateLazy_AnyValueParameters(producer IAnyProducer, options *LazyAnyValueOptions) error {
	return nil
}

func validateLazy_ListValueParameters(producer IListProducer, options *LazyListValueOptions) error {
	return nil
}

func validateLazy_NumberValueParameters(producer INumberProducer) error {
	return nil
}

func validateLazy_StringValueParameters(producer IStringProducer, options *LazyStringValueOptions) error {
	return nil
}

