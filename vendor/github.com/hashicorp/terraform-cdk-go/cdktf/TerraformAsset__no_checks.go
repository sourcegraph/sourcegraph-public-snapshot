// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func validateTerraformAsset_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_TerraformAsset) validateSetAssetHashParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_TerraformAsset) validateSetTypeParameters(val AssetType) error {
	return nil
}

func validateNewTerraformAssetParameters(scope constructs.Construct, id *string, config *TerraformAssetConfig) error {
	return nil
}

