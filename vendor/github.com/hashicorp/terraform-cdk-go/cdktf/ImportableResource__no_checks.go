// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func (i *jsiiProxy_ImportableResource) validateAddOverrideParameters(path *string, value interface{}) error {
	return nil
}

func (i *jsiiProxy_ImportableResource) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	return nil
}

func validateImportableResource_IsConstructParameters(x interface{}) error {
	return nil
}

func validateImportableResource_IsTerraformElementParameters(x interface{}) error {
	return nil
}

func validateNewImportableResourceParameters(scope constructs.Construct, name *string, config IImportableConfig) error {
	return nil
}

