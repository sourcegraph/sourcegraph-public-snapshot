// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (d *jsiiProxy_DefaultTokenResolver) validateResolveListParameters(xs *[]*string, context IResolveContext) error {
	return nil
}

func (d *jsiiProxy_DefaultTokenResolver) validateResolveMapParameters(xs *map[string]interface{}, context IResolveContext) error {
	return nil
}

func (d *jsiiProxy_DefaultTokenResolver) validateResolveNumberListParameters(xs *[]*float64, context IResolveContext) error {
	return nil
}

func (d *jsiiProxy_DefaultTokenResolver) validateResolveStringParameters(fragments TokenizedStringFragments, context IResolveContext) error {
	return nil
}

func (d *jsiiProxy_DefaultTokenResolver) validateResolveTokenParameters(t IResolvable, context IResolveContext, postProcessor IPostProcessor) error {
	return nil
}

func validateNewDefaultTokenResolverParameters(concat IFragmentConcatenator) error {
	return nil
}

