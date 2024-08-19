// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"
)

func (i *jsiiProxy_IPostProcessor) validatePostProcessParameters(input interface{}, context IResolveContext) error {
	if input == nil {
		return fmt.Errorf("parameter input is required, but nil was provided")
	}

	if context == nil {
		return fmt.Errorf("parameter context is required, but nil was provided")
	}

	return nil
}

