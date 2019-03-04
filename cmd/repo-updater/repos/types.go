package repos

import (
	"reflect"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// Repo represents a source code repository stored in Sourcegraph.
type Repo struct {
	// The internal Sourcegraph repo ID.
	ID uint32
	// Name is the name for this repository (e.g., "github.com/user/repo").
	//
	// Previously, this was called RepoURI.
	Name string
	// Description is a brief description of the repository.
	Description string
	// Language is the primary programming language used in this repository.
	Language string
	// Fork is whether this repository is a fork of another repository.
	Fork bool
	// Enabled is whether the repository is enabled. Disabled repositories are
	// not accessible by users (except site admins).
	Enabled bool
	// Archived is whether the repository has been archived.
	Archived bool
	// CreatedAt is when this repository was created on Sourcegraph.
	CreatedAt time.Time
	// UpdatedAt is when this repository's metadata was last updated on Sourcegraph.
	UpdatedAt time.Time
	// DeletedAt is when this repository was soft-deleted from Sourcegraph.
	DeletedAt time.Time
	// ExternalRepo identifies this repository by its ID on the external service where it resides (and the external
	// service itself).
	ExternalRepo api.ExternalRepoSpec
	// Sources identifies all the repo sources this Repo belongs to.
	Sources map[string]*SourceInfo
	// Metadata contains the raw source code host JSON metadata.
	Metadata interface{}
}

// A SourceInfo represents a source a Repo belongs to (such as an external service).
type SourceInfo struct {
	ID       string
	CloneURL string
}

// CloneURLs returns all the clone URLs this repo is clonable from.
func (r *Repo) CloneURLs() []string {
	urls := make([]string, 0, len(r.Sources))
	for _, src := range r.Sources {
		if src != nil && src.CloneURL != "" {
			urls = append(urls, src.CloneURL)
		}
	}
	return urls
}

// IsDeleted returns true if the repo is deleted.
func (r *Repo) IsDeleted() bool { return !r.DeletedAt.IsZero() }

// Update updates Repo r with the fields from the given newer Repo n,
// returning true if modified.
func (r *Repo) Update(n *Repo) (modified bool) {
	if !r.ExternalRepo.Equal(&n.ExternalRepo) && r.Name != n.Name {
		return false
	}

	if r.Name != n.Name {
		r.Name, modified = n.Name, true
	}

	if r.Description != n.Description {
		r.Description, modified = n.Description, true
	}

	if r.Language != n.Language {
		r.Language, modified = n.Language, true
	}

	if n.ExternalRepo != (api.ExternalRepoSpec{}) &&
		!r.ExternalRepo.Equal(&n.ExternalRepo) {
		r.ExternalRepo, modified = n.ExternalRepo, true
	}

	if r.Archived != n.Archived {
		r.Archived, modified = n.Archived, true
	}

	if r.Fork != n.Fork {
		r.Fork, modified = n.Fork, true
	}

	if !reflect.DeepEqual(r.Sources, n.Sources) {
		r.Sources, modified = n.Sources, true
	}

	if !reflect.DeepEqual(r.Metadata, n.Metadata) {
		r.Metadata, modified = n.Metadata, true
	}

	return modified
}

// Clone returns a clone of the given repo.
func (r *Repo) Clone() *Repo {
	clone := *r
	return &clone
}

// Apply applies the given functional options to the Repo.
func (r *Repo) Apply(opts ...func(*Repo)) {
	if r == nil {
		return
	}

	for _, opt := range opts {
		opt(r)
	}
}

// With returns a clone of the given repo with the given functional options applied.
func (r *Repo) With(opts ...func(*Repo)) *Repo {
	clone := r.Clone()
	clone.Apply(opts...)
	return clone
}

// Repos is an utility type with convenience methods for operating on lists of Repos.
type Repos []*Repo

// Names returns the list of names from all Repos.
func (rs Repos) Names() []string {
	names := make([]string, len(rs))
	for i := range rs {
		names[i] = rs[i].Name
	}
	return names
}

func (rs Repos) Len() int {
	return len(rs)
}

func (rs Repos) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

func (rs Repos) Less(i, j int) bool {
	if rs[i].Name == rs[j].Name {
		return rs[i].ExternalRepo.Compare(rs[j].ExternalRepo) == -1
	}
	return rs[i].Name < rs[j].Name
}

func (rs *Repos) Concat(others ...Repos) {
	for _, o := range others {
		*rs = append(*rs, o...)
	}
}

func (rs Repos) Clone() Repos {
	o := make(Repos, 0, len(rs))
	for _, r := range rs {
		o = append(o, r.Clone())
	}
	return o
}

// Apply applies the given functional options to the Repo.
func (rs Repos) Apply(opts ...func(*Repo)) {
	for _, r := range rs {
		r.Apply(opts...)
	}
}
