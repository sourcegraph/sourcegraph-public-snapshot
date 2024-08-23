// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type StackManifest struct {
	// Experimental.
	Annotations *[]*StackAnnotation `field:"required" json:"annotations" yaml:"annotations"`
	// Experimental.
	ConstructPath *string `field:"required" json:"constructPath" yaml:"constructPath"`
	// Experimental.
	Dependencies *[]*string `field:"required" json:"dependencies" yaml:"dependencies"`
	// Experimental.
	Name *string `field:"required" json:"name" yaml:"name"`
	// Experimental.
	StackMetadataPath *string `field:"required" json:"stackMetadataPath" yaml:"stackMetadataPath"`
	// Experimental.
	SynthesizedStackPath *string `field:"required" json:"synthesizedStackPath" yaml:"synthesizedStackPath"`
	// Experimental.
	WorkingDirectory *string `field:"required" json:"workingDirectory" yaml:"workingDirectory"`
}

