package graphql

import (
	"sort"
	"strings"
)

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

	FileMatches map[string]bool
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

func (r *Repository) SearchResultPaths() (list fileMatchPathList) {
	var files []string
	for f := range r.FileMatches {
		files = append(files, f)
	}
	sort.Strings(files)
	return fileMatchPathList(files)
}

type fileMatchPathList []string

func (f fileMatchPathList) String() string { return strings.Join(f, " ") }
