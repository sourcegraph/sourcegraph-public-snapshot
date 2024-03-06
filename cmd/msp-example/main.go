package main

import (
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"

	"github.com/sourcegraph/sourcegraph/cmd/msp-example/internal/example"
)

func main() {
	println("Try to Deliver")
	println("Service Account Token Creator")
	println("Cloud Deploy Releaser")
	println("Storage Legacy Bucket Reader")
	runtime.Start[example.Config](example.Service{})
}
