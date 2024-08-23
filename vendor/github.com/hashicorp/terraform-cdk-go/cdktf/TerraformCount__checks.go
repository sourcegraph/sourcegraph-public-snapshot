// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"
)

func validateTerraformCount_IsTerraformCountParameters(x interface{}) error {
	if x == nil {
		return fmt.Errorf("parameter x is required, but nil was provided")
	}

	return nil
}

func validateTerraformCount_OfParameters(count *float64) error {
	if count == nil {
		return fmt.Errorf("parameter count is required, but nil was provided")
	}

	return nil
}

