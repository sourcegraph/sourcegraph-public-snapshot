package main

import (
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"

	"github.com/sourcegraph/sourcegraph/cmd/msp-example/internal/example"
)

func main() {
	println("why you no run hello world")
	println("plzz run")
	println("_dab_")
	runtime.Start[example.Config](example.Service{})
}
