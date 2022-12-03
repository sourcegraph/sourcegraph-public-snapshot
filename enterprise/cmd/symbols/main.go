package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/symbols/shared"
	shared_symbols "github.com/sourcegraph/sourcegraph/cmd/symbols/shared"
	shared_symbols_enterprise "github.com/sourcegraph/sourcegraph/enterprise/cmd/symbols/shared"
)

func main() {
	shared_symbols.Main(shared_symbols_enterprise.CreateSetup(shared.CtagsConfig, shared.RepositoryFetcherConfig))
}
