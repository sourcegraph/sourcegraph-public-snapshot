// Command frontend is the enterprise frontend program.
package main

import (
	shared "github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry"
	shared_enterprise "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

func main() {
	env.Lock()
	env.HandleHelpFlag()

	shared.Main(shared_enterprise.EnterpriseSetupHook)
}

func init() {
	oobmigration.ReturnEnterpriseMigrations = true
}
