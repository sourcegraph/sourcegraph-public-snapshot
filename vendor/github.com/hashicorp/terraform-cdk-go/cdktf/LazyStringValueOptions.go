// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Options for creating a lazy string token.
// Experimental.
type LazyStringValueOptions struct {
	// Use the given name as a display hint.
	// Default: - No hint.
	//
	// Experimental.
	DisplayHint *string `field:"optional" json:"displayHint" yaml:"displayHint"`
}

