// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (m *jsiiProxy_Manifest) validateForStackParameters(stack TerraformStack) error {
	return nil
}

func validateNewManifestParameters(version *string, outdir *string, hclOutput *bool) error {
	return nil
}

