package graphql

import "github.com/sourcegraph/src-cli/internal/batches/util"

const RepositoryFieldsFragment = `
fragment repositoryFields on Repository {
    id
    name
    externalRepository {
        serviceType
    }
    defaultBranch {
        name
        target {
            oid
        }
    }
    commit(rev: $rev) @include(if:$queryCommit) {
        oid
    }
}
`

type Target struct {
	OID string
}

type Branch struct {
	Name   string
	Target Target
}

type Repository struct {
	ID                 string
	Name               string
	ExternalRepository struct{ ServiceType string }

	DefaultBranch *Branch

	Commit Target
	// Branch is populated by resolveRepositoryNameAndBranch with the queried
	// branch's name and the contents of the Commit property.
	Branch Branch

	FileMatches map[string]bool
}

func (r *Repository) HasBranch() bool {
	return r.DefaultBranch != nil || (r.Commit.OID != "" && r.Branch.Name != "")
}

func (r *Repository) BaseRef() string {
	if r.Branch.Name != "" {
		return util.EnsureRefPrefix(r.Branch.Name)
	}

	return util.EnsureRefPrefix(r.DefaultBranch.Name)
}

func (r *Repository) Rev() string {
	if r.Branch.Target.OID != "" {
		return r.Branch.Target.OID
	}

	return r.DefaultBranch.Target.OID
}
