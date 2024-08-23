// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"
)

func (d *jsiiProxy_DefaultTokenResolver) validateResolveListParameters(xs *[]*string, context IResolveContext) error {
	if xs == nil {
		return fmt.Errorf("parameter xs is required, but nil was provided")
	}

	if context == nil {
		return fmt.Errorf("parameter context is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DefaultTokenResolver) validateResolveMapParameters(xs *map[string]interface{}, context IResolveContext) error {
	if xs == nil {
		return fmt.Errorf("parameter xs is required, but nil was provided")
	}

	if context == nil {
		return fmt.Errorf("parameter context is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DefaultTokenResolver) validateResolveNumberListParameters(xs *[]*float64, context IResolveContext) error {
	if xs == nil {
		return fmt.Errorf("parameter xs is required, but nil was provided")
	}

	if context == nil {
		return fmt.Errorf("parameter context is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DefaultTokenResolver) validateResolveStringParameters(fragments TokenizedStringFragments, context IResolveContext) error {
	if fragments == nil {
		return fmt.Errorf("parameter fragments is required, but nil was provided")
	}

	if context == nil {
		return fmt.Errorf("parameter context is required, but nil was provided")
	}

	return nil
}

func (d *jsiiProxy_DefaultTokenResolver) validateResolveTokenParameters(t IResolvable, context IResolveContext, postProcessor IPostProcessor) error {
	if t == nil {
		return fmt.Errorf("parameter t is required, but nil was provided")
	}

	if context == nil {
		return fmt.Errorf("parameter context is required, but nil was provided")
	}

	if postProcessor == nil {
		return fmt.Errorf("parameter postProcessor is required, but nil was provided")
	}

	return nil
}

func validateNewDefaultTokenResolverParameters(concat IFragmentConcatenator) error {
	if concat == nil {
		return fmt.Errorf("parameter concat is required, but nil was provided")
	}

	return nil
}

