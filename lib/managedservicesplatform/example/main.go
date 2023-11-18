package main

import (
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/example/internal/example"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/service"
)

func main() {
	service.Run[example.Config](example.Service{})
}
