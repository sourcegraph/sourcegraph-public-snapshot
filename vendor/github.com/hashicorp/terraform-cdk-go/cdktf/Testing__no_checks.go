// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func validateTesting_AppParameters(options *TestingAppConfig) error {
	return nil
}

func validateTesting_EnableFutureFlagsParameters(app App) error {
	return nil
}

func validateTesting_FakeCdktfJsonPathParameters(app App) error {
	return nil
}

func validateTesting_FullSynthParameters(stack TerraformStack) error {
	return nil
}

func validateTesting_RenderConstructTreeParameters(construct constructs.IConstruct) error {
	return nil
}

func validateTesting_StubVersionParameters(app App) error {
	return nil
}

func validateTesting_SynthParameters(stack TerraformStack) error {
	return nil
}

func validateTesting_SynthHclParameters(stack TerraformStack) error {
	return nil
}

func validateTesting_SynthScopeParameters(fn IScopeCallback) error {
	return nil
}

func validateTesting_ToBeValidTerraformParameters(received *string) error {
	return nil
}

func validateTesting_ToHaveDataSourceParameters(received *string, resourceType *string) error {
	return nil
}

func validateTesting_ToHaveDataSourceWithPropertiesParameters(received *string, resourceType *string) error {
	return nil
}

func validateTesting_ToHaveProviderParameters(received *string, resourceType *string) error {
	return nil
}

func validateTesting_ToHaveProviderWithPropertiesParameters(received *string, resourceType *string) error {
	return nil
}

func validateTesting_ToHaveResourceParameters(received *string, resourceType *string) error {
	return nil
}

func validateTesting_ToHaveResourceWithPropertiesParameters(received *string, resourceType *string) error {
	return nil
}

