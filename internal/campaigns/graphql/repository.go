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

type Repository struct {
	ID                 string
	Name               string
	URL                string
	ExternalRepository struct{ ServiceType string }
	DefaultBranch      *struct {
		Name   string
		Target struct{ OID string }
	}
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
