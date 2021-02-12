package graphql

import (
	"crypto/sha256"
	"encoding/base64"
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
	URL                string
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
		return ensurePrefix(r.Branch.Name)
	}

	return r.DefaultBranch.Name
}

func ensurePrefix(rev string) string {
	if strings.HasPrefix(rev, "refs/heads/") {
		return rev
	}
	return "refs/heads/" + rev
}

func (r *Repository) Rev() string {
	if r.Branch.Target.OID != "" {
		return r.Branch.Target.OID
	}

	return r.DefaultBranch.Target.OID
}

func (r *Repository) Slug() string {
	return strings.ReplaceAll(r.Name, "/", "-") + "-" + r.Rev()
}

func (r *Repository) SlugForPath(path string) string {
	name := r.Name
	if path != "" {
		// Since path can contain os.PathSeparator or other characters that
		// don't translate well between Windows and Unix systems, we hash it.
		hash := sha256.Sum256([]byte(path))
		name = name + "-" + base64.RawURLEncoding.EncodeToString(hash[:32])
	}
	return strings.ReplaceAll(name, "/", "-") + "-" + r.Rev()
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
