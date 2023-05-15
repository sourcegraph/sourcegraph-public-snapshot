package main

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/shared"
	"github.com/sourcegraph/sourcegraph/internal/service/svcmain"
)

func main() {
	svcmain.SingleServiceMain(shared.Service, svcmain.Config{}, true, false)
}
