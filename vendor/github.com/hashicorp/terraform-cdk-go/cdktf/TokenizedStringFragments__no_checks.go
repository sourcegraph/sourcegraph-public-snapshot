// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (t *jsiiProxy_TokenizedStringFragments) validateAddEscapeParameters(kind *string) error {
	return nil
}

func (t *jsiiProxy_TokenizedStringFragments) validateAddIntrinsicParameters(value interface{}) error {
	return nil
}

func (t *jsiiProxy_TokenizedStringFragments) validateAddLiteralParameters(lit interface{}) error {
	return nil
}

func (t *jsiiProxy_TokenizedStringFragments) validateAddTokenParameters(token IResolvable) error {
	return nil
}

func (t *jsiiProxy_TokenizedStringFragments) validateConcatParameters(other TokenizedStringFragments) error {
	return nil
}

func (t *jsiiProxy_TokenizedStringFragments) validateJoinParameters(concat IFragmentConcatenator) error {
	return nil
}

func (t *jsiiProxy_TokenizedStringFragments) validateMapTokensParameters(context IResolveContext) error {
	return nil
}

