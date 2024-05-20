package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/service"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

func main() {
	runtime.Start(&service.Service{})
}
