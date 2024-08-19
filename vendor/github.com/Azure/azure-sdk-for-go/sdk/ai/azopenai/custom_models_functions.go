// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package azopenai

import (
	"encoding/json"
	"errors"
)

// These models and their code are used for the old style of function calling. They were auto-generated
// in the past (and won't change since they're an evolutionary dead-end at this point) but some older
// and supported models can only use this variant.

// ChatCompletionsOptionsFunctionCall - Controls how the model responds to function calls. "none" means the model does not
// call a function, and responds to the end-user. "auto" means the model can pick between an end-user or calling a
// function. Specifying a particular function via {"name": "my_function"} forces the model to call that function. "none" is
// the default when no functions are present. "auto" is the default if functions
// are present.
type ChatCompletionsOptionsFunctionCall struct {
	// IsFunction is true if Value refers to a function name.
	IsFunction bool

	// Value is one of:
	// - "auto", meaning the model can pick between an end-user or calling a function
	// - "none", meaning the model does not call a function,
	// - name of a function, in which case [IsFunction] should be set to true.
	Value *string
}

// MarshalJSON implements the json.Marshaller interface for type ChatCompletionsOptionsFunctionCall.
func (c ChatCompletionsOptionsFunctionCall) MarshalJSON() ([]byte, error) {
	if c.IsFunction {
		if c.Value == nil {
			return nil, errors.New("the Value should be the function name to call, not nil")
		}

		return json.Marshal(map[string]string{"name": *c.Value})
	}

	return json.Marshal(c.Value)
}
