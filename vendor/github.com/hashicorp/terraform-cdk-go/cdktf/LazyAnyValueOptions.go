// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Options for creating lazy untyped tokens.
// Experimental.
type LazyAnyValueOptions struct {
	// Use the given name as a display hint.
	// Default: - No hint.
	//
	// Experimental.
	DisplayHint *string `field:"optional" json:"displayHint" yaml:"displayHint"`
	// If the produced value is an array and it is empty, return 'undefined' instead.
	// Default: false.
	//
	// Experimental.
	OmitEmptyArray *bool `field:"optional" json:"omitEmptyArray" yaml:"omitEmptyArray"`
}

