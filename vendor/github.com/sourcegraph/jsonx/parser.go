// This file was ported from https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts,
// which is licensed as follows:
//
// Copyright (c) Microsoft Corporation. All rights reserved. Licensed under the MIT License.

package jsonx

import (
	"encoding/json"
)

// ParseOptions specifies options for JSON parsing.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L629
type ParseOptions struct {
	Comments       bool // allow comments (`//` and `/* ... */`)
	TrailingCommas bool // allow trailing commas in objects and arrays
}

// Parse the given text and returns the standard JSON representation of it,
// excluding the extensions supported by this package (such as comments and
// trailing commas).
//
// On invalid input, the parser tries to be as fault tolerant as possible,
// but still return a result. Callers should check the errors list to see
// if the input was valid.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L638
func Parse(text string, options ParseOptions) ([]byte, []ParseErrorCode) {
	var currentProperty struct {
		name  string
		valid bool
	}
	type parent struct {
		array  *[]interface{}
		object map[string]interface{}
	}
	currentParent := parent{array: &[]interface{}{}}
	previousParents := []parent{}

	onValue := func(value interface{}) {
		if currentParent.array != nil {
			*currentParent.array = append(*currentParent.array, value)
		} else if currentProperty.valid {
			currentParent.object[currentProperty.name] = value
		} else {
			panic("unreachable")
		}
	}

	var errors []ParseErrorCode
	visitor := Visitor{
		OnObjectBegin: func(offset, length int) {
			object := map[string]interface{}{}
			onValue(object)
			previousParents = append(previousParents, currentParent)
			currentParent = parent{object: object}
			currentProperty.name = ""
			currentProperty.valid = false
		},
		OnObjectProperty: func(property string, offset, length int) {
			currentProperty.name = property
			currentProperty.valid = true
		},
		OnObjectEnd: func(offset, length int) {
			currentParent = previousParents[len(previousParents)-1]
			previousParents = previousParents[:len(previousParents)-1]
		},
		OnArrayBegin: func(offset, length int) {
			array := &[]interface{}{}
			onValue(array)
			previousParents = append(previousParents, currentParent)
			currentParent = parent{array: array}
			currentProperty.name = ""
			currentProperty.valid = false
		},
		OnArrayEnd: func(offset, length int) {
			currentParent = previousParents[len(previousParents)-1]
			previousParents = previousParents[:len(previousParents)-1]
		},
		OnLiteralValue: func(value interface{}, offset, length int) {
			onValue(value)
		},
		OnError: func(errorCode ParseErrorCode, offset, length int) {
			errors = append(errors, errorCode)
		},
	}
	Walk(text, options, visitor)

	if len(*currentParent.array) == 0 {
		return nil, errors
	}
	data, err := json.Marshal((*currentParent.array)[0])
	if err != nil {
		panic(err) // should never happen
	}
	return data, errors
}

// A ParseErrorCode is a category of error that can occur while parsing a
// JSON document.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L593
type ParseErrorCode int

// Parse error codes
const (
	InvalidSymbol ParseErrorCode = iota
	InvalidNumberFormat
	PropertyNameExpected
	ValueExpected
	ColonExpected
	CommaExpected
	CloseBraceExpected
	CloseBracketExpected
	EndOfFileExpected
)
