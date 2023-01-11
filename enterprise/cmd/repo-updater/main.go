package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	enterprise_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater/shared"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

func main() {
	env.Lock()
	env.HandleHelpFlag()

	shared.Main(enterprise_shared.EnterpriseInit)
}
