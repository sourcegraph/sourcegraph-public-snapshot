// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"
)

func (t *jsiiProxy_TokenizedStringFragments) validateAddEscapeParameters(kind *string) error {
	if kind == nil {
		return fmt.Errorf("parameter kind is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TokenizedStringFragments) validateAddIntrinsicParameters(value interface{}) error {
	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TokenizedStringFragments) validateAddLiteralParameters(lit interface{}) error {
	if lit == nil {
		return fmt.Errorf("parameter lit is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TokenizedStringFragments) validateAddTokenParameters(token IResolvable) error {
	if token == nil {
		return fmt.Errorf("parameter token is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TokenizedStringFragments) validateConcatParameters(other TokenizedStringFragments) error {
	if other == nil {
		return fmt.Errorf("parameter other is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TokenizedStringFragments) validateJoinParameters(concat IFragmentConcatenator) error {
	if concat == nil {
		return fmt.Errorf("parameter concat is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TokenizedStringFragments) validateMapTokensParameters(context IResolveContext) error {
	if context == nil {
		return fmt.Errorf("parameter context is required, but nil was provided")
	}

	return nil
}

