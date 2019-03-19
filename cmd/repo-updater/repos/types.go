package repos

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

// An ExternalService is defines a Source that yields Repos.
type ExternalService struct {
	ID          int64
	Kind        string
	DisplayName string
	Config      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}

// URN returns a unique resource identifier of this external service,
// used as the key in a repo's Sources map as well as the SourceInfo ID.
func (e *ExternalService) URN() string {
	return "extsvc:" + strconv.FormatInt(e.ID, 10)
}

// IsDeleted returns true if the external service is deleted.
func (e *ExternalService) IsDeleted() bool { return !e.DeletedAt.IsZero() }

// Update updates ExternalService r with the fields from the given newer ExternalService n,
// returning true if modified.
func (e *ExternalService) Update(n *ExternalService) (modified bool) {
	if e.ID != n.ID {
		return false
	}

	if !strings.EqualFold(e.Kind, n.Kind) {
		e.Kind, modified = strings.ToUpper(n.Kind), true
	}

	if e.DisplayName != n.DisplayName {
		e.DisplayName, modified = n.DisplayName, true
	}

	if e.Config != n.Config {
		e.Config, modified = n.Config, true
	}

	if !e.UpdatedAt.Equal(n.UpdatedAt) {
		e.UpdatedAt, modified = n.UpdatedAt, true
	}

	if !e.DeletedAt.Equal(n.DeletedAt) {
		e.DeletedAt, modified = n.DeletedAt, true
	}

	return modified
}

// ExcludeGithubRepos changes the configuration of a Github external service to exclude the
// given repos from being synced.
func (e *ExternalService) ExcludeGithubRepos(rs ...*Repo) error {
	return e.config("github", func(v interface{}) error {
		c := v.(*schema.GitHubConnection)
		set := make(map[string]bool, len(c.Exclude)*2)
		for _, ex := range c.Exclude {
			if ex.Id != "" {
				set[ex.Id] = true
			}

			if ex.Name != "" {
				set[strings.ToLower(ex.Name)] = true
			}
		}

		for _, r := range rs {
			if r.ExternalRepo.ServiceType != "github" {
				continue
			}

			id := r.ExternalRepo.ID
			name := strings.ToLower(r.Name)

			if !set[name] && !set[id] {
				c.Exclude = append(c.Exclude, &schema.Exclude{
					Name: r.Name,
					Id:   id,
				})

				if id != "" {
					set[id] = true
				}

				if name != "" {
					set[name] = true
				}
			}
		}

		return nil
	})
}

// IncludeGithubRepos changes the configuration of a Github external service to explicitly enlist the
// given repos to be synced.
func (e *ExternalService) IncludeGithubRepos(rs ...*Repo) error {
	return e.config("github", func(v interface{}) error {
		c := v.(*schema.GitHubConnection)

		set := make(map[string]bool, len(c.Repos))
		for _, name := range c.Repos {
			set[strings.ToLower(name)] = true
		}

		for _, r := range rs {
			if r.ExternalRepo.ServiceType != "github" {
				continue
			}

			if name := strings.ToLower(r.Name); !set[name] {
				c.Repos = append(c.Repos, r.Name)
				set[name] = true
			}
		}

		return nil
	})
}

func (e *ExternalService) config(kind string, opt func(c interface{}) error) error {
	if strings.ToLower(e.Kind) != kind {
		return fmt.Errorf("config: unexpected external service kind %q", e.Kind)
	}

	var c interface{}
	switch kind {
	case "github":
		c = new(schema.GitHubConnection)
	default:
		panic("not implemented")
	}

	if err := jsonc.Unmarshal(e.Config, c); err != nil {
		return fmt.Errorf("external service id=%d config unmarshaling error: %s", e.ID, err)
	}

	if err := opt(c); err != nil {
		return errors.Wrap(err, "configure")
	}

	edited, err := jsonc.Edit(e.Config, c)
	if err != nil {
		return errors.Wrap(err, "edit")
	}

	e.Config = edited

	return nil
}

// Clone returns a clone of the given external service.
func (e *ExternalService) Clone() *ExternalService {
	clone := *e
	return &clone
}

// Apply applies the given functional options to the ExternalService.
func (e *ExternalService) Apply(opts ...func(*ExternalService)) {
	if e == nil {
		return
	}

	for _, opt := range opts {
		opt(e)
	}
}

// With returns a clone of the given repo with the given functional options applied.
func (e *ExternalService) With(opts ...func(*ExternalService)) *ExternalService {
	clone := e.Clone()
	clone.Apply(opts...)
	return clone
}

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

// ExternalServiceID returns the ID of the external service this
// SourceInfo refers to.
func (i SourceInfo) ExternalServiceID() int64 {
	ps := strings.SplitN(i.ID, ":", 2)
	if len(ps) != 2 {
		return 0
	}

	id, _ := strconv.ParseInt(ps[1], 10, 64)
	return id
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
	if rs[i].ID != rs[j].ID {
		return rs[i].ID < rs[j].ID
	}
	if rs[i].Name != rs[j].Name {
		return rs[i].Name < rs[j].Name
	}
	return rs[i].ExternalRepo.Compare(rs[j].ExternalRepo) == -1
}

// Concat adds the given Repos to the end of rs.
func (rs *Repos) Concat(others ...Repos) {
	for _, o := range others {
		*rs = append(*rs, o...)
	}
}

// Clone returns a clone of Repos.
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

// Filter returns all the Repos that match the given predicate.
func (rs Repos) Filter(pred func(*Repo) bool) (fs Repos) {
	for _, r := range rs {
		if pred(r) {
			fs = append(fs, r)
		}
	}
	return fs
}

// ExternalServices is an utility type with
// convenience methods for operating on lists of ExternalServices.
type ExternalServices []*ExternalService

// DisplayNames returns the list of display names from all ExternalServices.
func (es ExternalServices) DisplayNames() []string {
	names := make([]string, len(es))
	for i := range es {
		names[i] = es[i].DisplayName
	}
	return names
}

// URNs returns the list of URNs from all ExternalServices.
func (es ExternalServices) URNs() []string {
	urns := make([]string, len(es))
	for i := range es {
		urns[i] = es[i].URN()
	}
	return urns
}

func (es ExternalServices) Len() int {
	return len(es)
}

func (es ExternalServices) Swap(i, j int) {
	es[i], es[j] = es[j], es[i]
}

func (es ExternalServices) Less(i, j int) bool {
	return es[i].ID < es[j].ID
}

// Clone returns a clone of the given external services.
func (es ExternalServices) Clone() ExternalServices {
	o := make(ExternalServices, 0, len(es))
	for _, r := range es {
		o = append(o, r.Clone())
	}
	return o
}

// Apply applies the given functional options to the ExternalService.
func (es ExternalServices) Apply(opts ...func(*ExternalService)) {
	for _, r := range es {
		r.Apply(opts...)
	}
}

// With returns a clone of the given external services with the given functional options applied.
func (es ExternalServices) With(opts ...func(*ExternalService)) ExternalServices {
	clone := es.Clone()
	clone.Apply(opts...)
	return clone
}

// Filter returns all the ExternalServices that match the given predicate.
func (es ExternalServices) Filter(pred func(*ExternalService) bool) (fs ExternalServices) {
	for _, e := range es {
		if pred(e) {
			fs = append(fs, e)
		}
	}
	return fs
}
