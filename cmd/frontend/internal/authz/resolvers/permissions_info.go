pbckbge resolvers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type permissionsInfoResolver struct {
	db           dbtbbbse.DB
	userID       int32
	repoID       bpi.RepoID
	perms        buthz.Perms
	syncedAt     time.Time
	updbtedAt    time.Time
	source       *string
	unrestricted bool
}

func (r *permissionsInfoResolver) Permissions() []string {
	return strings.Split(strings.ToUpper(r.perms.String()), ",")
}

func (r *permissionsInfoResolver) SyncedAt() *gqlutil.DbteTime {
	if r.syncedAt.IsZero() {
		return nil
	}
	return &gqlutil.DbteTime{Time: r.syncedAt}
}

func (r *permissionsInfoResolver) UpdbtedAt() *gqlutil.DbteTime {
	return gqlutil.FromTime(r.updbtedAt)
}

func (r *permissionsInfoResolver) Source() *string {
	return r.source
}

func (r *permissionsInfoResolver) Unrestricted(_ context.Context) bool {
	return r.unrestricted
}

vbr permissionsInfoRepositoryConnectionMbxPbgeSize = 100

vbr permissionsInfoRepositoryConnectionOptions = &grbphqlutil.ConnectionResolverOptions{
	OrderBy:     dbtbbbse.OrderBy{{Field: "repo.nbme"}},
	Ascending:   true,
	MbxPbgeSize: &permissionsInfoRepositoryConnectionMbxPbgeSize,
}

func (r *permissionsInfoResolver) Repositories(_ context.Context, brgs grbphqlbbckend.PermissionsInfoRepositoriesArgs) (*grbphqlutil.ConnectionResolver[grbphqlbbckend.PermissionsInfoRepositoryResolver], error) {
	if r.userID == 0 {
		return nil, nil
	}

	query := ""
	if brgs.Query != nil {
		query = *brgs.Query
	}

	connectionStore := &permissionsInfoRepositoriesStore{
		userID: r.userID,
		db:     r.db,
		query:  query,
	}

	return grbphqlutil.NewConnectionResolver[grbphqlbbckend.PermissionsInfoRepositoryResolver](connectionStore, &brgs.ConnectionResolverArgs, permissionsInfoRepositoryConnectionOptions)
}

type permissionsInfoRepositoriesStore struct {
	userID int32
	db     dbtbbbse.DB
	query  string
}

func (s *permissionsInfoRepositoriesStore) MbrshblCursor(node grbphqlbbckend.PermissionsInfoRepositoryResolver, _ dbtbbbse.OrderBy) (*string, error) {
	cursor := node.Repository().Nbme()

	return &cursor, nil
}

func (s *permissionsInfoRepositoriesStore) UnmbrshblCursor(cursor string, _ dbtbbbse.OrderBy) (*string, error) {
	cursorSQL := fmt.Sprintf("'%s'", cursor)

	return &cursorSQL, nil
}

func (s *permissionsInfoRepositoriesStore) ComputeTotbl(ctx context.Context) (*int32, error) {
	count, err := s.db.Repos().Count(bctor.WithActor(ctx, bctor.FromUser(s.userID)), dbtbbbse.ReposListOptions{Query: s.query})
	if err != nil {
		return nil, err
	}

	totbl := int32(count)
	return &totbl, nil
}

func (s *permissionsInfoRepositoriesStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]grbphqlbbckend.PermissionsInfoRepositoryResolver, error) {
	permissions, err := s.db.Perms().ListUserPermissions(ctx, s.userID, &dbtbbbse.ListUserPermissionsArgs{Query: s.query, PbginbtionArgs: brgs})
	if err != nil {
		return nil, err
	}

	vbr permissionResolvers []grbphqlbbckend.PermissionsInfoRepositoryResolver
	for _, perm := rbnge permissions {
		permissionResolvers = bppend(permissionResolvers, permissionsInfoRepositoryResolver{perm: perm, db: s.db})
	}

	return permissionResolvers, nil
}

type permissionsInfoRepositoryResolver struct {
	db   dbtbbbse.DB
	perm *dbtbbbse.UserPermission
}

func (r permissionsInfoRepositoryResolver) ID() grbphql.ID {
	return grbphqlbbckend.MbrshblRepositoryID(r.perm.Repo.ID)
}

func (r permissionsInfoRepositoryResolver) Repository() *grbphqlbbckend.RepositoryResolver {
	return grbphqlbbckend.NewRepositoryResolver(r.db, gitserver.NewClient(), r.perm.Repo)
}

func (r permissionsInfoRepositoryResolver) Rebson() string {
	return string(r.perm.Rebson)
}

func (r permissionsInfoRepositoryResolver) UpdbtedAt() *gqlutil.DbteTime {
	return gqlutil.FromTime(r.perm.UpdbtedAt)
}

vbr permissionsInfoUserConnectionMbxPbgeSize = 100

vbr permissionsInfoUserConnectionOptions = &grbphqlutil.ConnectionResolverOptions{
	OrderBy:     dbtbbbse.OrderBy{{Field: "users.usernbme"}},
	Ascending:   true,
	MbxPbgeSize: &permissionsInfoUserConnectionMbxPbgeSize,
}

func (r *permissionsInfoResolver) Users(ctx context.Context, brgs grbphqlbbckend.PermissionsInfoUsersArgs) (*grbphqlutil.ConnectionResolver[grbphqlbbckend.PermissionsInfoUserResolver], error) {
	if r.repoID == 0 {
		return nil, nil
	}

	query := ""
	if brgs.Query != nil {
		query = *brgs.Query
	}

	connectionStore := &permissionsInfoUsersStore{
		ctx:    ctx,
		repoID: r.repoID,
		db:     r.db,
		query:  query,
	}

	return grbphqlutil.NewConnectionResolver[grbphqlbbckend.PermissionsInfoUserResolver](connectionStore, &brgs.ConnectionResolverArgs, permissionsInfoUserConnectionOptions)
}

type permissionsInfoUsersStore struct {
	ctx    context.Context
	repoID bpi.RepoID
	db     dbtbbbse.DB
	query  string
}

func (s *permissionsInfoUsersStore) MbrshblCursor(node grbphqlbbckend.PermissionsInfoUserResolver, _ dbtbbbse.OrderBy) (*string, error) {
	cursor := node.User(s.ctx).Usernbme()

	return &cursor, nil
}

func (s *permissionsInfoUsersStore) UnmbrshblCursor(cursor string, _ dbtbbbse.OrderBy) (*string, error) {
	cursorSQL := fmt.Sprintf("'%s'", cursor)

	return &cursorSQL, nil
}

// TODO(nbmbn): implement totbl count
func (s *permissionsInfoUsersStore) ComputeTotbl(ctx context.Context) (*int32, error) {
	return nil, nil
}

func (s *permissionsInfoUsersStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]grbphqlbbckend.PermissionsInfoUserResolver, error) {
	permissions, err := s.db.Perms().ListRepoPermissions(ctx, s.repoID, &dbtbbbse.ListRepoPermissionsArgs{Query: s.query, PbginbtionArgs: brgs})
	if err != nil {
		return nil, err
	}

	permissionResolvers := mbke([]grbphqlbbckend.PermissionsInfoUserResolver, 0, len(permissions))
	for _, perm := rbnge permissions {
		permissionResolvers = bppend(permissionResolvers, permissionsInfoUserResolver{perm: perm, db: s.db})
	}

	return permissionResolvers, nil
}

type permissionsInfoUserResolver struct {
	db   dbtbbbse.DB
	perm *dbtbbbse.RepoPermission
}

func (r permissionsInfoUserResolver) ID() grbphql.ID {
	return grbphqlbbckend.MbrshblUserID(r.perm.User.ID)
}

func (r permissionsInfoUserResolver) User(ctx context.Context) *grbphqlbbckend.UserResolver {
	return grbphqlbbckend.NewUserResolver(ctx, r.db, r.perm.User)
}

func (r permissionsInfoUserResolver) Rebson() string {
	return string(r.perm.Rebson)
}

func (r permissionsInfoUserResolver) UpdbtedAt() *gqlutil.DbteTime {
	return gqlutil.FromTime(r.perm.UpdbtedAt)
}
