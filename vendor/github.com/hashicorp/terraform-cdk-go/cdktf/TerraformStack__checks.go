// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"

	"github.com/aws/constructs-go/constructs/v10"
)

func (t *jsiiProxy_TerraformStack) validateAddDependencyParameters(dependency TerraformStack) error {
	if dependency == nil {
		return fmt.Errorf("parameter dependency is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TerraformStack) validateAddOverrideParameters(path *string, value interface{}) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TerraformStack) validateAllocateLogicalIdParameters(tfElement interface{}) error {
	if tfElement == nil {
		return fmt.Errorf("parameter tfElement is required, but nil was provided")
	}
	switch tfElement.(type) {
	case constructs.Node:
		// ok
	case TerraformElement:
		// ok
	default:
		if !_jsii_.IsAnonymousProxy(tfElement) {
			return fmt.Errorf("parameter tfElement must be one of the allowed types: constructs.Node, TerraformElement; received %#v (a %T)", tfElement, tfElement)
		}
	}

	return nil
}

func (t *jsiiProxy_TerraformStack) validateDependsOnParameters(stack TerraformStack) error {
	if stack == nil {
		return fmt.Errorf("parameter stack is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TerraformStack) validateGetLogicalIdParameters(tfElement interface{}) error {
	if tfElement == nil {
		return fmt.Errorf("parameter tfElement is required, but nil was provided")
	}
	switch tfElement.(type) {
	case constructs.Node:
		// ok
	case TerraformElement:
		// ok
	default:
		if !_jsii_.IsAnonymousProxy(tfElement) {
			return fmt.Errorf("parameter tfElement must be one of the allowed types: constructs.Node, TerraformElement; received %#v (a %T)", tfElement, tfElement)
		}
	}

	return nil
}

func (t *jsiiProxy_TerraformStack) validateRegisterIncomingCrossStackReferenceParameters(fromStack TerraformStack) error {
	if fromStack == nil {
		return fmt.Errorf("parameter fromStack is required, but nil was provided")
	}

	return nil
}

func (t *jsiiProxy_TerraformStack) validateRegisterOutgoingCrossStackReferenceParameters(identifier *string) error {
	if identifier == nil {
		return fmt.Errorf("parameter identifier is required, but nil was provided")
	}

	return nil
}

func validateTerraformStack_IsConstructParameters(x interface{}) error {
	if x == nil {
		return fmt.Errorf("parameter x is required, but nil was provided")
	}

	return nil
}

func validateTerraformStack_IsStackParameters(x interface{}) error {
	if x == nil {
		return fmt.Errorf("parameter x is required, but nil was provided")
	}

	return nil
}

func validateTerraformStack_OfParameters(construct constructs.IConstruct) error {
	if construct == nil {
		return fmt.Errorf("parameter construct is required, but nil was provided")
	}

	return nil
}

func (j *jsiiProxy_TerraformStack) validateSetDependenciesParameters(val *[]TerraformStack) error {
	if val == nil {
		return fmt.Errorf("parameter val is required, but nil was provided")
	}

	return nil
}

func (j *jsiiProxy_TerraformStack) validateSetMoveTargetsParameters(val TerraformResourceTargets) error {
	if val == nil {
		return fmt.Errorf("parameter val is required, but nil was provided")
	}

	return nil
}

func (j *jsiiProxy_TerraformStack) validateSetSynthesizerParameters(val IStackSynthesizer) error {
	if val == nil {
		return fmt.Errorf("parameter val is required, but nil was provided")
	}

	return nil
}

func validateNewTerraformStackParameters(scope constructs.Construct, id *string) error {
	if scope == nil {
		return fmt.Errorf("parameter scope is required, but nil was provided")
	}

	if id == nil {
		return fmt.Errorf("parameter id is required, but nil was provided")
	}

	return nil
}

