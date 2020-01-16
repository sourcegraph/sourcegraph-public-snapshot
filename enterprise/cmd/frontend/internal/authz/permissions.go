package authz

import (
	"time"

	"github.com/RoaringBitmap/roaring"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
)

// PermType is the object type of the user permissions.
type PermType string

// The list of available user permission types.
const (
	PermRepos PermType = "repos"
)

// ProviderType is the type of provider implementation for the permissions.
type ProviderType string

// The list of available provider types.
const (
	ProviderBitbucketServer ProviderType = bitbucketserver.ServiceType
	ProviderSourcegraph     ProviderType = "sourcegraph"
)

// UserPermissions are the permissions of a user to perform an action
// on the given set of object IDs of the defined type that is scoped by
// the provider.
type UserPermissions struct {
	UserID    int32
	Perm      authz.Perms
	Type      PermType
	IDs       *roaring.Bitmap
	Provider  ProviderType
	UpdatedAt time.Time
}

// Expired returns true if these UserPermissions have elapsed the given ttl.
func (p *UserPermissions) Expired(ttl time.Duration, now time.Time) bool {
	return !now.Before(p.UpdatedAt.Add(ttl))
}

// AuthorizedRepos returns the intersection of the given repository IDs with
// the authorized IDs.
func (p *UserPermissions) AuthorizedRepos(repos []*types.Repo) []authz.RepoPerms {
	if p.Type != PermRepos {
		return nil
	}

	perms := make([]authz.RepoPerms, 0, len(repos))
	for _, r := range repos {
		if r.ID != 0 && p.IDs != nil && p.IDs.Contains(uint32(r.ID)) {
			perms = append(perms, authz.RepoPerms{Repo: r, Perms: p.Perm})
		}
	}
	return perms
}

// TracingFields returns tracing fields for the opentracing log.
func (p *UserPermissions) TracingFields() []otlog.Field {
	fs := []otlog.Field{
		otlog.Int32("UserPermissions.UserID", p.UserID),
		otlog.String("UserPermissions.Perm", string(p.Perm)),
		otlog.String("UserPermissions.Type", string(p.Type)),
		otlog.String("UserPermissions.Provider", string(p.Provider)),
	}

	if p.IDs != nil {
		fs = append(fs,
			otlog.Uint64("UserPermissions.IDs.Count", p.IDs.GetCardinality()),
			otlog.String("UserPermissions.UpdatedAt", p.UpdatedAt.String()),
		)
	}

	return fs
}

// RepoPermissions declares which users have access to a given repository, as
// defined by the permissions provider (either Bitbucket Server or Sourcegraph).
type RepoPermissions struct {
	RepoID    int32
	Perm      authz.Perms
	UserIDs   *roaring.Bitmap
	Provider  ProviderType
	UpdatedAt time.Time
}

// Expired returns true if these RepoPermissions have elapsed the given ttl.
func (p *RepoPermissions) Expired(ttl time.Duration, now time.Time) bool {
	return !now.Before(p.UpdatedAt.Add(ttl))
}

// AuthorizedUsers returns the intersection of the given user IDs with
// the authorized IDs.
func (p *RepoPermissions) AuthorizedUsers(users []*types.User) []authz.RepoPerms {
	perms := make([]authz.RepoPerms, 0, len(users))
	for _, u := range users {
		if u.ID != 0 && p.UserIDs != nil && p.UserIDs.Contains(uint32(u.ID)) {
			perms = append(perms, authz.RepoPerms{
				Repo: &types.Repo{
					ID: api.RepoID(p.RepoID),
				},
				Perms: p.Perm,
			})
		}
	}
	return perms
}

// TracingFields returns tracing fields for the opentracing log.
func (p *RepoPermissions) TracingFields() []otlog.Field {
	fs := []otlog.Field{
		otlog.Int32("RepoPermissions.RepoID", p.RepoID),
		otlog.String("RepoPermissions.Perm", string(p.Perm)),
		otlog.String("RepoPermissions.Provider", string(p.Provider)),
	}

	if p.UserIDs != nil {
		fs = append(fs,
			otlog.Uint64("RepoPermissions.UserIDs.Count", p.UserIDs.GetCardinality()),
			otlog.String("RepoPermissions.UpdatedAt", p.UpdatedAt.String()),
		)
	}

	return fs
}

// UserPendingPermissions defines permissions that a not-yet-created user has to
// perform on a given set of object IDs. Not-yet-created users may exist on the
// code host but not yet in Sourcegraph. `BindID` is used to map this stub user
// to an actual user when the actual user is created; it can either be a username
// or email.
type UserPendingPermissions struct {
	ID        int32
	BindID    string
	Perm      authz.Perms
	Type      PermType
	IDs       *roaring.Bitmap
	UpdatedAt time.Time
}

// Expired returns true if these UserPendingPermissions have elapsed the given ttl.
func (p *UserPendingPermissions) Expired(ttl time.Duration, now time.Time) bool {
	return !now.Before(p.UpdatedAt.Add(ttl))
}

// AuthorizedRepos returns the intersection of the given repository IDs with
// the authorized IDs.
func (p *UserPendingPermissions) AuthorizedRepos(repos []*types.Repo) []authz.RepoPerms {
	if p.Type != PermRepos {
		return nil
	}

	perms := make([]authz.RepoPerms, 0, len(repos))
	for _, r := range repos {
		if r.ID != 0 && p.IDs != nil && p.IDs.Contains(uint32(r.ID)) {
			perms = append(perms, authz.RepoPerms{Repo: r, Perms: p.Perm})
		}
	}
	return perms
}

// TracingFields returns tracing fields for the opentracing log.
func (p *UserPendingPermissions) TracingFields() []otlog.Field {
	fs := []otlog.Field{
		otlog.Int32("UserPendingPermissions.ID", p.ID),
		otlog.String("UserPendingPermissions.BindID", p.BindID),
		otlog.String("UserPendingPermissions.Perm", string(p.Perm)),
		otlog.String("UserPendingPermissions.Type", string(p.Type)),
	}

	if p.IDs != nil {
		fs = append(fs,
			otlog.Uint64("UserPendingPermissions.IDs.Count", p.IDs.GetCardinality()),
			otlog.String("UserPendingPermissions.UpdatedAt", p.UpdatedAt.String()),
		)
	}

	return fs
}
