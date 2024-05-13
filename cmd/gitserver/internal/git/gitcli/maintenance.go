package gitcli

import "github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"

func (g *gitCLIBackend) Maintenance() git.MaintenanceBackend {
	return g
}
