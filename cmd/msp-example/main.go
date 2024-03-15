package main

import (
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"

	"github.com/sourcegraph/sourcegraph/cmd/msp-example/internal/example"
)

func main() {
	println("testing, testing, 1...2...testing")
	println("Adding a code change")
	println("Code change again")
	runtime.Start[example.Config](example.Service{})
}
