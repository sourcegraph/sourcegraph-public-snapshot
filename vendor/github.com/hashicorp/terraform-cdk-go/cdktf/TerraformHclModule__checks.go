// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"

	"github.com/aws/constructs-go/constructs/v10"
)

func (t *jsiiProxy_TerraformHclModule) validateAddOverrideParameters(path *string, value interface{}) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TerraformHclModule) validateAddProviderParameters(provider interface{}) error {
	if provider == nil {
		return fmt.Errorf("parameter provider is required, but nil was provided")
	}
	switch provider.(type) {
	case TerraformProvider:
		// ok
	case *TerraformModuleProvider:
		provider := provider.(*TerraformModuleProvider)
		if err := _jsii_.ValidateStruct(provider, func() string { return "parameter provider" }); err != nil {
			return err
		}
	case TerraformModuleProvider:
		provider_ := provider.(TerraformModuleProvider)
		provider := &provider_
		if err := _jsii_.ValidateStruct(provider, func() string { return "parameter provider" }); err != nil {
			return err
		}
	default:
		if !_jsii_.IsAnonymousProxy(provider) {
			return fmt.Errorf("parameter provider must be one of the allowed types: TerraformProvider, *TerraformModuleProvider; received %#v (a %T)", provider, provider)
		}
	}

	return nil
}

func (t *jsiiProxy_TerraformHclModule) validateGetParameters(output *string) error {
	if output == nil {
		return fmt.Errorf("parameter output is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TerraformHclModule) validateGetBooleanParameters(output *string) error {
	if output == nil {
		return fmt.Errorf("parameter output is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TerraformHclModule) validateGetListParameters(output *string) error {
	if output == nil {
		return fmt.Errorf("parameter output is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TerraformHclModule) validateGetNumberParameters(output *string) error {
	if output == nil {
		return fmt.Errorf("parameter output is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TerraformHclModule) validateGetStringParameters(output *string) error {
	if output == nil {
		return fmt.Errorf("parameter output is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TerraformHclModule) validateInterpolationForOutputParameters(moduleOutput *string) error {
	if moduleOutput == nil {
		return fmt.Errorf("parameter moduleOutput is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TerraformHclModule) validateOverrideLogicalIdParameters(newLogicalId *string) error {
	if newLogicalId == nil {
		return fmt.Errorf("parameter newLogicalId is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TerraformHclModule) validateSetParameters(variable *string, value interface{}) error {
	if variable == nil {
		return fmt.Errorf("parameter variable is required, but nil was provided")
	}

	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateTerraformHclModule_IsConstructParameters(x interface{}) error {
	if x == nil {
		return fmt.Errorf("parameter x is required, but nil was provided")
	}

	return nil
}

func validateTerraformHclModule_IsTerraformElementParameters(x interface{}) error {
	if x == nil {
		return fmt.Errorf("parameter x is required, but nil was provided")
	}

	return nil
}

func validateNewTerraformHclModuleParameters(scope constructs.Construct, id *string, options *TerraformHclModuleConfig) error {
	if scope == nil {
		return fmt.Errorf("parameter scope is required, but nil was provided")
	}

	if id == nil {
		return fmt.Errorf("parameter id is required, but nil was provided")
	}

	if options == nil {
		return fmt.Errorf("parameter options is required, but nil was provided")
	}
	if err := _jsii_.ValidateStruct(options, func() string { return "parameter options" }); err != nil {
		return err
	}

	return nil
}

