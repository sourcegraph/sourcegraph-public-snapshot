// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type StackAnnotation struct {
	// Experimental.
	ConstructPath *string `field:"required" json:"constructPath" yaml:"constructPath"`
	// Experimental.
	Level AnnotationMetadataEntryType `field:"required" json:"level" yaml:"level"`
	// Experimental.
	Message *string `field:"required" json:"message" yaml:"message"`
	// Experimental.
	Stacktrace *[]*string `field:"optional" json:"stacktrace" yaml:"stacktrace"`
}

