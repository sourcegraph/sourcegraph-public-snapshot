package repos

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// A Changeset of an existing Repo.
type Changeset struct {
	Title   string
	Body    string
	HeadRef string
	BaseRef string

	*campaigns.Changeset
	*Repo
}

// IsOutdated returns true when the attributes of the nested
// campaigns.Changeset do not match the attributes (title, body, ...) set on
// the Changeset.
func (c *Changeset) IsOutdated() (bool, error) {
	currentTitle, err := c.Changeset.Title()
	if err != nil {
		return false, err
	}

	if currentTitle != c.Title {
		return true, nil
	}

	currentBody, err := c.Changeset.Body()
	if err != nil {
		return false, err
	}

	if currentBody != c.Body {
		return true, nil
	}

	currentBaseRef, err := c.Changeset.BaseRef()
	if err != nil {
		return false, err
	}

	if git.EnsureRefPrefix(currentBaseRef) != git.EnsureRefPrefix(c.BaseRef) {
		return true, nil
	}

	return false, nil
}

// Repo represents a source code repository stored in Sourcegraph.
type Repo struct {
	// The internal Sourcegraph repo ID.
	ID api.RepoID
	// Name is the name for this repository (e.g., "github.com/user/repo"). It
	// is the same as URI, unless the user configures a non-default
	// repositoryPathPattern.
	//
	// Previously, this was called RepoURI.
	Name string
	// URI is the full name for this repository (e.g.,
	// "github.com/user/repo"). See the documentation for the Name field.
	URI string
	// Description is a brief description of the repository.
	Description string
	// Fork is whether this repository is a fork of another repository.
	Fork bool
	// Archived is whether the repository has been archived.
	Archived bool
	// Private is whether the repository is private.
	Private bool
	// Cloned is whether the repository is cloned.
	Cloned bool
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
	// The key is a URN created by extsvc.URN
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
	ps := strings.SplitN(i.ID, ":", 3)
	if len(ps) != 3 {
		return -1
	}

	id, err := strconv.ParseInt(ps[2], 10, 64)
	if err != nil {
		return -1
	}

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

// ExternalServiceIDs returns the IDs of the external services this
// repo belongs to.
func (r *Repo) ExternalServiceIDs() []int64 {
	ids := make([]int64, 0, len(r.Sources))
	for _, src := range r.Sources {
		ids = append(ids, src.ExternalServiceID())
	}
	return ids
}

// IsDeleted returns true if the repo is deleted.
func (r *Repo) IsDeleted() bool { return !r.DeletedAt.IsZero() }

// Update updates Repo r with the fields from the given newer Repo n,
// returning true if modified.
func (r *Repo) Update(n *Repo) (modified bool) {
	if r.Name != n.Name {
		r.Name, modified = n.Name, true
	}

	if r.URI != n.URI {
		r.URI, modified = n.URI, true
	}

	if r.Description != n.Description {
		r.Description, modified = n.Description, true
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

	if r.Private != n.Private {
		r.Private, modified = n.Private, true
	}

	if !reflect.DeepEqual(r.Sources, n.Sources) {
		r.Sources, modified = n.Sources, true
	}

	// As a special case, we clear out the value of ViewerPermission for GitHub repos as
	// the value is dependent on the token used to fetch it. We don't want to store this in the DB as it will
	// flip flop as we fetch the same repo from different external services.
	switch x := n.Metadata.(type) {
	case *github.Repository:
		cp := *x
		cp.ViewerPermission = ""
		n = n.With(func(clone *Repo) {
			// Repo.Clone does not currently clone metadata for any types as they could contain hard to clone
			// items such as maps. However, we know that copying github.Repository is safe as it only contains values.
			clone.Metadata = &cp
		})
	}

	if !reflect.DeepEqual(r.Metadata, n.Metadata) {
		r.Metadata, modified = n.Metadata, true
	}

	return modified
}

// Clone returns a clone of the given repo.
func (r *Repo) Clone() *Repo {
	if r == nil {
		return nil
	}
	clone := *r
	if r.Sources != nil {
		clone.Sources = make(map[string]*SourceInfo, len(r.Sources))
		for k, v := range r.Sources {
			clone.Sources[k] = v
		}
	}
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

// Less compares Repos by the important fields (fields with constraints in our
// DB). Additionally it will compare on Sources to give a deterministic order
// on repos returned from a sourcer.
//
// NewDiff relies on Less to deterministically decide on the order to merge
// repositories, as well as which repository to keep on conflicts.
//
// Context on using other fields such as timestamps to order/resolve
// conflicts: We only want to rely on values that have constraints in our
// database. Timestamps have the following downsides:
//
//   - We need to assume the upstream codehost has reasonable values for them
//   - Not all codehosts set them to relevant values (eg gitolite or other)
//   - They could change often for codehosts that do set them.
func (r *Repo) Less(s *Repo) bool {
	if r.ID != s.ID {
		return r.ID < s.ID
	}
	if r.Name != s.Name {
		return r.Name < s.Name
	}
	if cmp := r.ExternalRepo.Compare(s.ExternalRepo); cmp != 0 {
		return cmp == -1
	}

	return sortedSliceLess(sourcesKeys(r.Sources), sourcesKeys(s.Sources))
}

func (r *Repo) String() string {
	eid := fmt.Sprintf("{%s %s %s}", r.ExternalRepo.ServiceID, r.ExternalRepo.ServiceType, r.ExternalRepo.ID)
	if r.IsDeleted() {
		return fmt.Sprintf("Repo{ID: %d, Name: %q, EID: %s, IsDeleted: true}", r.ID, r.Name, eid)
	}
	return fmt.Sprintf("Repo{ID: %d, Name: %q, EID: %s}", r.ID, r.Name, eid)
}

func sourcesKeys(m map[string]*SourceInfo) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedSliceLess returns true if a < b
func sortedSliceLess(a, b []string) bool {
	for i, v := range a {
		if i == len(b) {
			return false
		}
		if v != b[i] {
			return v < b[i]
		}
	}
	return len(a) != len(b)
}

// pick deterministically chooses between a and b a repo to keep and
// discard. It is used when resolving conflicts on sourced repositories.
func pick(a *Repo, b *Repo) (keep, discard *Repo) {
	if a.Less(b) {
		return a, b
	}
	return b, a
}

// Repos is an utility type with convenience methods for operating on lists of Repos.
type Repos []*Repo

// IDs returns the list of ids from all Repos.
func (rs Repos) IDs() []api.RepoID {
	ids := make([]api.RepoID, len(rs))
	for i := range rs {
		ids[i] = rs[i].ID
	}
	return ids
}

// Names returns the list of names from all Repos.
func (rs Repos) Names() []string {
	names := make([]string, len(rs))
	for i := range rs {
		names[i] = rs[i].Name
	}
	return names
}

// NamesSummary caps the number of repos to 20 when composing a space-separated list string.
// Used in logging statements.
func (rs Repos) NamesSummary() string {
	if len(rs) > 20 {
		return strings.Join(rs[:20].Names(), " ") + "..."
	}
	return strings.Join(rs.Names(), " ")
}

// Kinds returns the unique set of kinds from all Repos.
func (rs Repos) Kinds() (kinds []string) {
	set := map[string]bool{}
	for _, r := range rs {
		kind := strings.ToUpper(r.ExternalRepo.ServiceType)
		if !set[kind] {
			kinds = append(kinds, kind)
			set[kind] = true
		}
	}
	return kinds
}

// ExternalRepos returns the list of set ExternalRepoSpecs from all Repos.
func (rs Repos) ExternalRepos() []api.ExternalRepoSpec {
	specs := make([]api.ExternalRepoSpec, 0, len(rs))
	for _, r := range rs {
		specs = append(specs, r.ExternalRepo)
	}
	return specs
}

// Sources returns a map of all the sources per repo id.
func (rs Repos) Sources() map[api.RepoID][]SourceInfo {
	sources := make(map[api.RepoID][]SourceInfo)
	for i := range rs {
		for _, info := range rs[i].Sources {
			sources[rs[i].ID] = append(sources[rs[i].ID], *info)
		}
	}

	return sources
}

func (rs Repos) Len() int {
	return len(rs)
}

func (rs Repos) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

func (rs Repos) Less(i, j int) bool {
	return rs[i].Less(rs[j])
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

// With returns a clone of the given repos with the given functional options applied.
func (rs Repos) With(opts ...func(*Repo)) Repos {
	clone := rs.Clone()
	clone.Apply(opts...)
	return clone
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

type externalServiceLister interface {
	ListExternalServices(context.Context, StoreListExternalServicesArgs) ([]*types.ExternalService, error)
}

// RateLimitSyncer syncs rate limits based on external service configuration
type RateLimitSyncer struct {
	registry      *ratelimit.Registry
	serviceLister externalServiceLister
	// How many services to fetch in each DB call
	limit int64
}

// NewRateLimitSyncer returns a new syncer
func NewRateLimitSyncer(registry *ratelimit.Registry, serviceLister externalServiceLister) *RateLimitSyncer {
	r := &RateLimitSyncer{
		registry:      registry,
		serviceLister: serviceLister,
		limit:         500,
	}
	return r
}

// SyncRateLimiters syncs all rate limiters using current config.
// We sync them all as we need to pick the most restrictive configured limit per code host
// and rate limits can be defined in multiple external services for the same host.
func (r *RateLimitSyncer) SyncRateLimiters(ctx context.Context) error {
	var cursor int64
	byURL := make(map[string]extsvc.RateLimitConfig)

	for {
		services, err := r.serviceLister.ListExternalServices(ctx, StoreListExternalServicesArgs{
			Cursor: cursor,
			Limit:  r.limit,
		})
		if err != nil {
			return errors.Wrap(err, "listing external services")
		}

		if len(services) == 0 {
			break
		}

		cursor = services[len(services)-1].ID

		for _, svc := range services {
			rlc, err := extsvc.ExtractRateLimitConfig(svc.Config, svc.Kind, svc.DisplayName)
			if err != nil {
				if _, ok := err.(extsvc.ErrRateLimitUnsupported); ok {
					continue
				}
				return errors.Wrap(err, "getting rate limit configuration")
			}

			current, ok := byURL[rlc.BaseURL]
			if !ok || (ok && current.IsDefault) {
				byURL[rlc.BaseURL] = rlc
				continue
			}
			// Use the lower limit, but a default value should not override
			// a limit that has been configured
			if rlc.Limit < current.Limit && !rlc.IsDefault {
				byURL[rlc.BaseURL] = rlc
			}
		}

		if len(services) < int(r.limit) {
			break
		}
	}

	for u, rl := range byURL {
		l := r.registry.Get(u)
		l.SetLimit(rl.Limit)
	}

	return nil
}
