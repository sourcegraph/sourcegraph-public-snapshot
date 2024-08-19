// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"

	"github.com/aws/constructs-go/constructs/v10"
)

func validateTesting_AppParameters(options *TestingAppConfig) error {
	if err := _jsii_.ValidateStruct(options, func() string { return "parameter options" }); err != nil {
		return err
	}

	return nil
}

func validateTesting_EnableFutureFlagsParameters(app App) error {
	if app == nil {
		return fmt.Errorf("parameter app is required, but nil was provided")
	}

	return nil
}

func validateTesting_FakeCdktfJsonPathParameters(app App) error {
	if app == nil {
		return fmt.Errorf("parameter app is required, but nil was provided")
	}

	return nil
}

func validateTesting_FullSynthParameters(stack TerraformStack) error {
	if stack == nil {
		return fmt.Errorf("parameter stack is required, but nil was provided")
	}

	return nil
}

func validateTesting_RenderConstructTreeParameters(construct constructs.IConstruct) error {
	if construct == nil {
		return fmt.Errorf("parameter construct is required, but nil was provided")
	}

	return nil
}

func validateTesting_StubVersionParameters(app App) error {
	if app == nil {
		return fmt.Errorf("parameter app is required, but nil was provided")
	}

	return nil
}

func validateTesting_SynthParameters(stack TerraformStack) error {
	if stack == nil {
		return fmt.Errorf("parameter stack is required, but nil was provided")
	}

	return nil
}

func validateTesting_SynthHclParameters(stack TerraformStack) error {
	if stack == nil {
		return fmt.Errorf("parameter stack is required, but nil was provided")
	}

	return nil
}

func validateTesting_SynthScopeParameters(fn IScopeCallback) error {
	if fn == nil {
		return fmt.Errorf("parameter fn is required, but nil was provided")
	}

	return nil
}

func validateTesting_ToBeValidTerraformParameters(received *string) error {
	if received == nil {
		return fmt.Errorf("parameter received is required, but nil was provided")
	}

	return nil
}

func validateTesting_ToHaveDataSourceParameters(received *string, resourceType *string) error {
	if received == nil {
		return fmt.Errorf("parameter received is required, but nil was provided")
	}

	if resourceType == nil {
		return fmt.Errorf("parameter resourceType is required, but nil was provided")
	}

	return nil
}

func validateTesting_ToHaveDataSourceWithPropertiesParameters(received *string, resourceType *string) error {
	if received == nil {
		return fmt.Errorf("parameter received is required, but nil was provided")
	}

	if resourceType == nil {
		return fmt.Errorf("parameter resourceType is required, but nil was provided")
	}

	return nil
}

func validateTesting_ToHaveProviderParameters(received *string, resourceType *string) error {
	if received == nil {
		return fmt.Errorf("parameter received is required, but nil was provided")
	}

	if resourceType == nil {
		return fmt.Errorf("parameter resourceType is required, but nil was provided")
	}

	return nil
}

func validateTesting_ToHaveProviderWithPropertiesParameters(received *string, resourceType *string) error {
	if received == nil {
		return fmt.Errorf("parameter received is required, but nil was provided")
	}

	if resourceType == nil {
		return fmt.Errorf("parameter resourceType is required, but nil was provided")
	}

	return nil
}

func validateTesting_ToHaveResourceParameters(received *string, resourceType *string) error {
	if received == nil {
		return fmt.Errorf("parameter received is required, but nil was provided")
	}

	if resourceType == nil {
		return fmt.Errorf("parameter resourceType is required, but nil was provided")
	}

	return nil
}

func validateTesting_ToHaveResourceWithPropertiesParameters(received *string, resourceType *string) error {
	if received == nil {
		return fmt.Errorf("parameter received is required, but nil was provided")
	}

	if resourceType == nil {
		return fmt.Errorf("parameter resourceType is required, but nil was provided")
	}

	return nil
}

