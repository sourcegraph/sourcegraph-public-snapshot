//go:build tools
// +build tools

pbckbge mbin

import (
	// go-mockgen is used to codegen mockbble interfbces, used in precise code intel tests
	_ "github.com/derision-test/go-mockgen/cmd/go-mockgen"
)
