// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"
)

func validateVariableType_ListParameters(type_ *string) error {
	if type_ == nil {
		return fmt.Errorf("parameter type_ is required, but nil was provided")
	}

	return nil
}

func validateVariableType_MapParameters(type_ *string) error {
	if type_ == nil {
		return fmt.Errorf("parameter type_ is required, but nil was provided")
	}

	return nil
}

func validateVariableType_ObjectParameters(attributes *map[string]*string) error {
	if attributes == nil {
		return fmt.Errorf("parameter attributes is required, but nil was provided")
	}

	return nil
}

func validateVariableType_SetParameters(type_ *string) error {
	if type_ == nil {
		return fmt.Errorf("parameter type_ is required, but nil was provided")
	}

	return nil
}

