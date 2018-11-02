//docker:user sourcegraph
//docker:cmd serve
//docker:env CONFIGURATION_MODE=server
//docker:env PUBLIC_REPO_REDIRECTS=true

// Postgres defaults for cluster deployments.
//docker:env PGDATABASE=sg
//docker:env PGHOST=pgsql
//docker:env PGPORT=5432
//docker:env PGSSLMODE=disable
//docker:env PGUSER=sg

// Note: All frontend code should be added to shared.Main, not here. See that
// function for details.

package main

import (
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assets"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/registry"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
)

func main() {
	shared.Main()
}
