// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Options for creating a lazy list token.
// Experimental.
type LazyListValueOptions struct {
	// Use the given name as a display hint.
	// Default: - No hint.
	//
	// Experimental.
	DisplayHint *string `field:"optional" json:"displayHint" yaml:"displayHint"`
	// If the produced list is empty, return 'undefined' instead.
	// Default: false.
	//
	// Experimental.
	OmitEmpty *bool `field:"optional" json:"omitEmpty" yaml:"omitEmpty"`
}

