// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"

	"github.com/aws/constructs-go/constructs/v10"
)

// Includes API for attaching annotations such as warning messages to constructs.
// Experimental.
type Annotations interface {
	// Adds an { "error": <message> } metadata entry to this construct.
	//
	// The toolkit will fail synthesis when errors are reported.
	// Experimental.
	AddError(message *string)
	// Adds an info metadata entry to this construct.
	//
	// The CLI will display the info message when apps are synthesized.
	// Experimental.
	AddInfo(message *string)
	// Adds a warning metadata entry to this construct.
	//
	// The CLI will display the warning when an app is synthesized.
	// In a future release the CLI might introduce a --strict flag which
	// will then fail the synthesis if it encounters a warning.
	// Experimental.
	AddWarning(message *string)
}

// The jsii proxy struct for Annotations
type jsiiProxy_Annotations struct {
	_ byte // padding
}

// Returns the annotations API for a construct scope.
// Experimental.
func Annotations_Of(scope constructs.IConstruct) Annotations {
	_init_.Initialize()

	if err := validateAnnotations_OfParameters(scope); err != nil {
		panic(err)
	}
	var returns Annotations

	_jsii_.StaticInvoke(
		"cdktf.Annotations",
		"of",
		[]interface{}{scope},
		&returns,
	)

	return returns
}

func (a *jsiiProxy_Annotations) AddError(message *string) {
	if err := a.validateAddErrorParameters(message); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		a,
		"addError",
		[]interface{}{message},
	)
}

func (a *jsiiProxy_Annotations) AddInfo(message *string) {
	if err := a.validateAddInfoParameters(message); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		a,
		"addInfo",
		[]interface{}{message},
	)
}

func (a *jsiiProxy_Annotations) AddWarning(message *string) {
	if err := a.validateAddWarningParameters(message); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		a,
		"addWarning",
		[]interface{}{message},
	)
}

