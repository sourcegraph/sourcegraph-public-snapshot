package main

import (
	"github.com/sourcegraph/sourcegraph/dev/linearhooks/internal/service"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

func main() {
	runtime.Start[service.Config](service.Service{})
}
