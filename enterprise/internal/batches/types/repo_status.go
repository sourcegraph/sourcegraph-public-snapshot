package types

import "github.com/sourcegraph/sourcegraph/internal/api"

type RepoStatus struct {
	RepoID  api.RepoID
	Commit  string
	Ignored bool
}
