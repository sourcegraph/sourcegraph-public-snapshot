package main

import (
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"

	"github.com/sourcegraph/sourcegraph/cmd/msp-example/internal/example"
)

func main() {
	println("Try to Deliver")
	println("Cloud Deploy Releaser")
	println("roles/storage.objectAdmin")
	runtime.Start[example.Config](example.Service{})
}
