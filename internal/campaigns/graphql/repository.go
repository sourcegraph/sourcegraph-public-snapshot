package graphql

import "strings"

const RepositoryFieldsFragment = `
fragment repositoryFields on Repository {
    id
    name
    url
    externalRepository {
        serviceType
    }
    defaultBranch {
        name
        target {
            oid
        }
    }
}
`

type Branch struct {
	Name   string
	Target struct{ OID string }
}

type Repository struct {
	ID                 string
	Name               string
	URL                string
	ExternalRepository struct{ ServiceType string }
	DefaultBranch      *Branch
}

func (r *Repository) BaseRef() string {
	return r.DefaultBranch.Name
}

func (r *Repository) Rev() string {
	return r.DefaultBranch.Target.OID
}

func (r *Repository) Slug() string {
	return strings.ReplaceAll(r.Name, "/", "-")
}
