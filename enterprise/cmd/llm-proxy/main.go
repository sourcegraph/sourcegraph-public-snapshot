package main

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/shared"
	"github.com/sourcegraph/sourcegraph/internal/service/svcmain"
)

func main() {
	svcmain.DeprecatedSingleServiceMain(shared.Service, svcmain.Config{}, true, false)
}
