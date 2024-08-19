// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"
)

func (i *jsiiProxy_IResolveContext) validateRegisterPostProcessorParameters(postProcessor IPostProcessor) error {
	if postProcessor == nil {
		return fmt.Errorf("parameter postProcessor is required, but nil was provided")
	}

	return nil
}

func (i *jsiiProxy_IResolveContext) validateResolveParameters(x interface{}) error {
	if x == nil {
		return fmt.Errorf("parameter x is required, but nil was provided")
	}

	return nil
}

