// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func validateVariableType_ListParameters(type_ *string) error {
	return nil
}

func validateVariableType_MapParameters(type_ *string) error {
	return nil
}

func validateVariableType_ObjectParameters(attributes *map[string]*string) error {
	return nil
}

func validateVariableType_SetParameters(type_ *string) error {
	return nil
}

