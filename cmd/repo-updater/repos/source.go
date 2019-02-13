package repos

import (
	"context"
)

// A Source yields repositories to be stored and analysed by Sourcegraph.
// Successive calls to its ListRepos method may yield different results.
type Source interface {
	ListRepos(context.Context) ([]*Repo, error)
}
