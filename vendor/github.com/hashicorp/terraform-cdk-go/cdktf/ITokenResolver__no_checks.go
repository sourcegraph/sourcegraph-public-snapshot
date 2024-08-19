// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (i *jsiiProxy_ITokenResolver) validateResolveListParameters(l *[]*string, context IResolveContext) error {
	return nil
}

func (i *jsiiProxy_ITokenResolver) validateResolveMapParameters(m *map[string]interface{}, context IResolveContext) error {
	return nil
}

func (i *jsiiProxy_ITokenResolver) validateResolveNumberListParameters(l *[]*float64, context IResolveContext) error {
	return nil
}

func (i *jsiiProxy_ITokenResolver) validateResolveStringParameters(s TokenizedStringFragments, context IResolveContext) error {
	return nil
}

func (i *jsiiProxy_ITokenResolver) validateResolveTokenParameters(t IResolvable, context IResolveContext, postProcessor IPostProcessor) error {
	return nil
}

