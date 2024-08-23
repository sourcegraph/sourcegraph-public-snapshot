// Package jsii contains the functionaility needed for jsii packages to
// initialize their dependencies and themselves. Users should never need to use this package
// directly. If you find you need to - please report a bug at
// https://github.com/aws/jsii/issues/new/choose
package jsii

import (
	_      "embed"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

//go:embed constructs-10.3.0.tgz
var tarball []byte

// Initialize loads the necessary packages in the @jsii/kernel to support the enclosing module.
// The implementation is idempotent (and hence safe to be called over and over).
func Initialize() {
	// Load this library into the kernel
	_jsii_.Load("constructs", "10.3.0", tarball)
}
