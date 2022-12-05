package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/symbols/shared"
	enterpriseshared "github.com/sourcegraph/sourcegraph/enterprise/cmd/symbols/shared"
)

func main() {
	shared.Main(enterpriseshared.CreateSetup(shared.CtagsConfig, shared.RepositoryFetcherConfig))
}
