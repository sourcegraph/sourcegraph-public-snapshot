package main

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph/enterprisecmd/nodbcmd"
)

func main() {
	nodbcmd.DeprecatedSingleServiceMainEnterprise(shared.Service)
}
