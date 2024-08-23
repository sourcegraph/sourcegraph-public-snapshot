package constructs

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

// Implement this interface in order for the construct to be able to validate itself.
type IValidation interface {
	// Validate the current construct.
	//
	// This method can be implemented by derived constructs in order to perform
	// validation logic. It is called on all constructs before synthesis.
	//
	// Returns: An array of validation error messages, or an empty array if there the construct is valid.
	Validate() *[]*string
}

// The jsii proxy for IValidation
type jsiiProxy_IValidation struct {
	_ byte // padding
}

func (i *jsiiProxy_IValidation) Validate() *[]*string {
	var returns *[]*string

	_jsii_.Invoke(
		i,
		"validate",
		nil, // no parameters
		&returns,
	)

	return returns
}

