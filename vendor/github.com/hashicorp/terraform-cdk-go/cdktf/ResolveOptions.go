// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	"github.com/aws/constructs-go/constructs/v10"
)

// Options to the resolve() operation.
//
// NOT the same as the ResolveContext; ResolveContext is exposed to Token
// implementors and resolution hooks, whereas this struct is just to bundle
// a number of things that would otherwise be arguments to resolve() in a
// readable way.
// Experimental.
type ResolveOptions struct {
	// The resolver to apply to any resolvable tokens found.
	// Experimental.
	Resolver ITokenResolver `field:"required" json:"resolver" yaml:"resolver"`
	// The scope from which resolution is performed.
	// Experimental.
	Scope constructs.IConstruct `field:"required" json:"scope" yaml:"scope"`
	// Whether the resolution is being executed during the prepare phase or not.
	// Default: false.
	//
	// Experimental.
	Preparing *bool `field:"optional" json:"preparing" yaml:"preparing"`
}

