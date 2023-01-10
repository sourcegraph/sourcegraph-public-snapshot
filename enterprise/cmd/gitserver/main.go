package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/shared"
	enterprise_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/gitserver/shared"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

func main() {
	env.Lock()
	env.HandleHelpFlag()

	shared.Main(enterprise_shared.EnterpriseInit)
}
