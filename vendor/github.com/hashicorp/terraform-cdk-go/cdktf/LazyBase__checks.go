// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"
)

func (l *jsiiProxy_LazyBase) validateAddPostProcessorParameters(postProcessor IPostProcessor) error {
	if postProcessor == nil {
		return fmt.Errorf("parameter postProcessor is required, but nil was provided")
	}

	return nil
}

func (l *jsiiProxy_LazyBase) validateResolveParameters(context IResolveContext) error {
	if context == nil {
		return fmt.Errorf("parameter context is required, but nil was provided")
	}

	return nil
}

func (l *jsiiProxy_LazyBase) validateResolveLazyParameters(context IResolveContext) error {
	if context == nil {
		return fmt.Errorf("parameter context is required, but nil was provided")
	}

	return nil
}

