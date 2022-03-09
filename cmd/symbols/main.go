// Command symbols is a service that serves code symbols (functions, variables, etc.) from a repository at a
// specific commit.
package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/symbols/shared"
)

func main() {
	shared.Main(shared.SetupSqlite)
}
