package main

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph/enterprisecmd/executorcmd"
)

func main() {
	executorcmd.DeprecatedSingleServiceMainEnterprise(shared.Service)
}
