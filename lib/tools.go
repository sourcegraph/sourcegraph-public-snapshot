//go:build tools
// +build tools

package main

import (
	// go-mockgen is used to codegen mockable interfaces, used in precise code intel tests
	"fmt"

	_ "github.com/derision-test/go-mockgen/cmd/go-mockgen"
)

func uhuh() {
	fmt.Println("test")
}
