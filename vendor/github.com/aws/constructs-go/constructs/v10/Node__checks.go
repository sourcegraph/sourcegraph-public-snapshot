//go:build !no_runtime_type_checking

package constructs

import (
	"fmt"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func (n *jsiiProxy_Node) validateAddMetadataParameters(type_ *string, data interface{}, options *MetadataOptions) error {
	if type_ == nil {
		return fmt.Errorf("parameter type_ is required, but nil was provided")
	}

	if data == nil {
		return fmt.Errorf("parameter data is required, but nil was provided")
	}

	if err := _jsii_.ValidateStruct(options, func() string { return "parameter options" }); err != nil {
		return err
	}

	return nil
}

func (n *jsiiProxy_Node) validateAddValidationParameters(validation IValidation) error {
	if validation == nil {
		return fmt.Errorf("parameter validation is required, but nil was provided")
	}

	return nil
}

func (n *jsiiProxy_Node) validateFindChildParameters(id *string) error {
	if id == nil {
		return fmt.Errorf("parameter id is required, but nil was provided")
	}

	return nil
}

func (n *jsiiProxy_Node) validateGetContextParameters(key *string) error {
	if key == nil {
		return fmt.Errorf("parameter key is required, but nil was provided")
	}

	return nil
}

func (n *jsiiProxy_Node) validateSetContextParameters(key *string, value interface{}) error {
	if key == nil {
		return fmt.Errorf("parameter key is required, but nil was provided")
	}

	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func (n *jsiiProxy_Node) validateTryFindChildParameters(id *string) error {
	if id == nil {
		return fmt.Errorf("parameter id is required, but nil was provided")
	}

	return nil
}

func (n *jsiiProxy_Node) validateTryGetContextParameters(key *string) error {
	if key == nil {
		return fmt.Errorf("parameter key is required, but nil was provided")
	}

	return nil
}

func (n *jsiiProxy_Node) validateTryRemoveChildParameters(childName *string) error {
	if childName == nil {
		return fmt.Errorf("parameter childName is required, but nil was provided")
	}

	return nil
}

func validateNode_OfParameters(construct IConstruct) error {
	if construct == nil {
		return fmt.Errorf("parameter construct is required, but nil was provided")
	}

	return nil
}

func validateNewNodeParameters(host Construct, scope IConstruct, id *string) error {
	if host == nil {
		return fmt.Errorf("parameter host is required, but nil was provided")
	}

	if scope == nil {
		return fmt.Errorf("parameter scope is required, but nil was provided")
	}

	if id == nil {
		return fmt.Errorf("parameter id is required, but nil was provided")
	}

	return nil
}

