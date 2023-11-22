package main

import (
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/service"

	"github.com/sourcegraph/sourcegraph/cmd/msp-example/internal/example"
)

func main() {
	service.Run[example.Config](example.Service{})
}
