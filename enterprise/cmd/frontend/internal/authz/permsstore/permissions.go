package permsstore

import (
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
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
	ProviderBitbucketServer ProviderType = "bitbucket_server"
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

func (p *UserPermissions) loadQuery() *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permsstore/permissions.go:UserPermissions.loadQuery
SELECT object_ids, updated_at
FROM user_permissions
WHERE user_id = %s
AND permission = %s
AND object_type = %s
AND provider = %s
`

	return sqlf.Sprintf(
		format,
		p.UserID,
		p.Perm.String(),
		p.Type,
		p.Provider,
	)
}

func (p *UserPermissions) upsertQuery() (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permsstore/permissions.go:UserPermissions.upsertQuery
INSERT INTO user_permissions
  (user_id, permission, object_type, object_ids, provider, updated_at)
VALUES
  (%s, %s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  user_permissions_perm_object_provider_unique
DO UPDATE SET
  object_ids = excluded.object_ids,
  updated_at = excluded.updated_at
`

	ids, err := p.IDs.ToBytes()
	if err != nil {
		return nil, err
	}

	if p.UpdatedAt.IsZero() {
		return nil, errors.New("UpdatedAt timestamp must be set")
	}

	return sqlf.Sprintf(
		format,
		p.UserID,
		p.Perm.String(),
		p.Type,
		ids,
		p.Provider,
		p.UpdatedAt.UTC(),
	), nil
}

// RepoPermissions are the permissions of the given set of object IDs of the
// defined type can perform an action on the repository that is scoped by the
// provider.
type RepoPermissions struct {
	RepoID    int32
	Perm      authz.Perms
	IDs       *roaring.Bitmap
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
		if u.ID != 0 && p.IDs != nil && p.IDs.Contains(uint32(u.ID)) {
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

	if p.IDs != nil {
		fs = append(fs,
			otlog.Uint64("RepoPermissions.IDs.Count", p.IDs.GetCardinality()),
			otlog.String("RepoPermissions.UpdatedAt", p.UpdatedAt.String()),
		)
	}

	return fs
}

// diffIDs returns diffs of user IDs.
func (p *RepoPermissions) diffIDs(ids *roaring.Bitmap) (toAdd, toRemove []uint32) {
	iter := p.IDs.Iterator()
	for iter.HasNext() {
		id := iter.Next()
		if ids.Contains(id) {
			continue
		}
		toAdd = append(toAdd, id)
	}
	iter = ids.Iterator()
	for iter.HasNext() {
		id := iter.Next()
		if p.IDs.Contains(id) {
			continue
		}
		toRemove = append(toRemove, id)
	}
	return toAdd, toRemove
}

func (p *RepoPermissions) loadQuery() *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permsstore/permissions.go:RepoPermissions.loadQuery
SELECT object_ids, updated_at
FROM repo_permissions
WHERE repo_id = %s
AND permission = %s
AND provider = %s
`

	return sqlf.Sprintf(
		format,
		p.RepoID,
		p.Perm.String(),
		p.Provider,
	)
}

func (p *RepoPermissions) upsertQuery() (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permsstore/permissions.go:RepoPermissions.upsertQuery
INSERT INTO repo_permissions
  (repo_id, permission, object_ids, provider, updated_at)
VALUES
  (%s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  repo_permissions_perm_provider_unique
DO UPDATE SET
  object_ids = excluded.object_ids,
  updated_at = excluded.updated_at
`

	ids, err := p.IDs.ToBytes()
	if err != nil {
		return nil, err
	}

	if p.UpdatedAt.IsZero() {
		return nil, errors.New("UpdatedAt timestamp must be set")
	}

	return sqlf.Sprintf(
		format,
		p.RepoID,
		p.Perm.String(),
		ids,
		p.Provider,
		p.UpdatedAt.UTC(),
	), nil
}

// PendingPermissions are the pending permissions of a stub user to perform
// an action on the given set of object IDs of the defined type. It is currently
// only scoped to Sourcegraph provider. The BindID can either be username or email.
type PendingPermissions struct {
	BindID    string
	Perm      authz.Perms
	Type      PermType
	IDs       *roaring.Bitmap
	UpdatedAt time.Time
}

// Expired returns true if these PendingPermissions have elapsed the given ttl.
func (p *PendingPermissions) Expired(ttl time.Duration, now time.Time) bool {
	return !now.Before(p.UpdatedAt.Add(ttl))
}

// AuthorizedRepos returns the intersection of the given repository IDs with
// the authorized IDs.
func (p *PendingPermissions) AuthorizedRepos(repos []*types.Repo) []authz.RepoPerms {
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
func (p *PendingPermissions) TracingFields() []otlog.Field {
	fs := []otlog.Field{
		otlog.String("PendingPermissions.BindID", p.BindID),
		otlog.String("PendingPermissions.Perm", string(p.Perm)),
		otlog.String("PendingPermissions.Type", string(p.Type)),
	}

	if p.IDs != nil {
		fs = append(fs,
			otlog.Uint64("PendingPermissions.IDs.Count", p.IDs.GetCardinality()),
			otlog.String("PendingPermissions.UpdatedAt", p.UpdatedAt.String()),
		)
	}

	return fs
}

func (p *PendingPermissions) loadQuery() *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permsstore/permissions.go:PendingPermissions.loadQuery
SELECT object_ids, updated_at
FROM user_pending_permissions
WHERE bind_id = %s
AND permission = %s
AND object_type = %s
`

	return sqlf.Sprintf(
		format,
		p.BindID,
		p.Perm.String(),
		p.Type,
	)
}

func (p *PendingPermissions) loadWithBindIDQuery() *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permsstore/permissions.go:PendingPermissions.loadWithBindIDQuery
SELECT bind_id, object_ids, updated_at
FROM user_pending_permissions
WHERE permission = %s
AND object_type = %s
`

	return sqlf.Sprintf(
		format,
		p.Perm.String(),
		p.Type,
	)
}

func (p *PendingPermissions) upsertQuery() (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permsstore/permissions.go:PendingPermissions.upsertQuery
INSERT INTO user_pending_permissions
  (bind_id, permission, object_type, object_ids, updated_at)
VALUES
  (%s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  user_pending_permissions_perm_object_unique
DO UPDATE SET
  object_ids = excluded.object_ids,
  updated_at = excluded.updated_at
`

	ids, err := p.IDs.ToBytes()
	if err != nil {
		return nil, err
	}

	if p.UpdatedAt.IsZero() {
		return nil, errors.New("UpdatedAt timestamp must be set")
	}

	return sqlf.Sprintf(
		format,
		p.BindID,
		p.Perm.String(),
		p.Type,
		ids,
		p.UpdatedAt.UTC(),
	), nil
}
