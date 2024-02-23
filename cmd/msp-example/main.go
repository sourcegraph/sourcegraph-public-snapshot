package main

import (
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"

	"github.com/sourcegraph/sourcegraph/cmd/msp-example/internal/example"
)

func main() {
	println("hello")
	println("world")
	println("banana")
	println(":jotter:")
	runtime.Start[example.Config](example.Service{})
}
