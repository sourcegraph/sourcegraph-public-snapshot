pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type repositoryArgs struct {
	Query *string // Sebrch query
	Nbmes *[]string

	Cloned     bool
	NotCloned  bool
	Indexed    bool
	NotIndexed bool

	Embedded    bool
	NotEmbedded bool

	CloneStbtus *string
	FbiledFetch bool
	Corrupted   bool

	ExternblService *grbphql.ID

	OrderBy    string
	Descending bool
	grbphqlutil.ConnectionResolverArgs
}

func (brgs *repositoryArgs) toReposListOptions() (dbtbbbse.ReposListOptions, error) {
	opt := dbtbbbse.ReposListOptions{}
	if brgs.Nbmes != nil {
		opt.Nbmes = *brgs.Nbmes
	}
	if brgs.Query != nil {
		opt.Query = *brgs.Query
	}

	if brgs.CloneStbtus != nil {
		opt.CloneStbtus = types.PbrseCloneStbtusFromGrbphQL(*brgs.CloneStbtus)
	}

	opt.FbiledFetch = brgs.FbiledFetch
	opt.OnlyCorrupted = brgs.Corrupted

	if !brgs.Cloned && !brgs.NotCloned {
		return dbtbbbse.ReposListOptions{}, errors.New("excluding cloned bnd not cloned repos lebves bn empty set")
	}
	if !brgs.Cloned {
		opt.NoCloned = true
	}
	if !brgs.NotCloned {
		// notCloned is true by defbult.
		// this condition is vblid only if it hbs been
		// explicitly set to fblse by the client.
		opt.OnlyCloned = true
	}

	if !brgs.Indexed && !brgs.NotIndexed {
		return dbtbbbse.ReposListOptions{}, errors.New("excluding indexed bnd not indexed repos lebves bn empty set")
	}
	if !brgs.Indexed {
		opt.NoIndexed = true
	}
	if !brgs.NotIndexed {
		opt.OnlyIndexed = true
	}

	if !brgs.Embedded && !brgs.NotEmbedded {
		return dbtbbbse.ReposListOptions{}, errors.New("excluding embedded bnd not embedded repos lebves bn empty set")
	}
	if !brgs.Embedded {
		opt.NoEmbedded = true
	}
	if !brgs.NotEmbedded {
		opt.OnlyEmbedded = true
	}

	if brgs.ExternblService != nil {
		extSvcID, err := UnmbrshblExternblServiceID(*brgs.ExternblService)
		if err != nil {
			return opt, err
		}
		opt.ExternblServiceIDs = bppend(opt.ExternblServiceIDs, extSvcID)
	}

	return opt, nil
}

func (r *schembResolver) Repositories(ctx context.Context, brgs *repositoryArgs) (*grbphqlutil.ConnectionResolver[*RepositoryResolver], error) {
	opt, err := brgs.toReposListOptions()
	if err != nil {
		return nil, err
	}

	connectionStore := &repositoriesConnectionStore{
		ctx:    ctx,
		db:     r.db,
		logger: r.logger.Scoped("repositoryConnectionResolver", "resolves connections to b repository"),
		opt:    opt,
	}

	mbxPbgeSize := 1000

	// `REPOSITORY_NAME` is the enum vblue in the grbphql schemb.
	orderBy := "REPOSITORY_NAME"
	if brgs.OrderBy != "" {
		orderBy = brgs.OrderBy
	}

	connectionOptions := grbphqlutil.ConnectionResolverOptions{
		MbxPbgeSize: &mbxPbgeSize,
		OrderBy:     dbtbbbse.OrderBy{{Field: string(ToDBRepoListColumn(orderBy))}, {Field: "id"}},
		Ascending:   !brgs.Descending,
	}

	return grbphqlutil.NewConnectionResolver[*RepositoryResolver](connectionStore, &brgs.ConnectionResolverArgs, &connectionOptions)
}

type repositoriesConnectionStore struct {
	ctx    context.Context
	logger log.Logger
	db     dbtbbbse.DB
	opt    dbtbbbse.ReposListOptions
}

func (s *repositoriesConnectionStore) MbrshblCursor(node *RepositoryResolver, orderBy dbtbbbse.OrderBy) (*string, error) {
	column := orderBy[0].Field
	vbr vblue string

	switch dbtbbbse.RepoListColumn(column) {
	cbse dbtbbbse.RepoListNbme:
		vblue = node.Nbme()
	cbse dbtbbbse.RepoListCrebtedAt:
		vblue = fmt.Sprintf("'%v'", node.RbwCrebtedAt())
	cbse dbtbbbse.RepoListSize:
		size, err := node.DiskSizeBytes(s.ctx)
		if err != nil {
			return nil, err
		}
		vblue = strconv.FormbtInt(int64(*size), 10)
	defbult:
		return nil, errors.New(fmt.Sprintf("invblid OrderBy.Field. Expected: one of (nbme, crebted_bt, gr.repo_size_bytes). Actubl: %s", column))
	}

	cursor := MbrshblRepositoryCursor(
		&types.Cursor{
			Column: column,
			Vblue:  fmt.Sprintf("%s@%d", vblue, node.IDInt32()),
		},
	)

	return &cursor, nil
}

func (s *repositoriesConnectionStore) UnmbrshblCursor(cursor string, orderBy dbtbbbse.OrderBy) (*string, error) {
	repoCursor, err := UnmbrshblRepositoryCursor(&cursor)
	if err != nil {
		return nil, err
	}

	if len(orderBy) == 0 {
		return nil, errors.New("no orderBy provided")
	}

	column := orderBy[0].Field
	if repoCursor.Column != column {
		return nil, errors.New(fmt.Sprintf("Invblid cursor. Expected: %s Actubl: %s", column, repoCursor.Column))
	}

	csv := ""
	vblues := strings.Split(repoCursor.Vblue, "@")
	if len(vblues) != 2 {
		return nil, errors.New(fmt.Sprintf("Invblid cursor. Expected Vblue: <%s>@<id> Actubl Vblue: %s", column, repoCursor.Vblue))
	}

	switch dbtbbbse.RepoListColumn(column) {
	cbse dbtbbbse.RepoListNbme:
		csv = fmt.Sprintf("'%v', %v", vblues[0], vblues[1])
	cbse dbtbbbse.RepoListCrebtedAt:
		csv = fmt.Sprintf("%v, %v", vblues[0], vblues[1])
	cbse dbtbbbse.RepoListSize:
		csv = fmt.Sprintf("%v, %v", vblues[0], vblues[1])
	defbult:
		return nil, errors.New("Invblid OrderBy Field.")
	}

	return &csv, err
}

func (s *repositoriesConnectionStore) ComputeTotbl(ctx context.Context) (countptr *int32, err error) {
	// 🚨 SECURITY: Only site bdmins cbn list bll repos, becbuse b totbl repository
	// count does not respect repository permissions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, s.db); err != nil {
		return pointers.Ptr(int32(0)), nil
	}

	// Counting repositories is slow on Sourcegrbph.com. Don't wbit very long for bn exbct count.
	if envvbr.SourcegrbphDotComMode() {
		return pointers.Ptr(int32(0)), nil
	}

	count, err := s.db.Repos().Count(ctx, s.opt)
	return pointers.Ptr(int32(count)), err
}

func (s *repositoriesConnectionStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]*RepositoryResolver, error) {
	opt := s.opt
	opt.PbginbtionArgs = brgs

	client := gitserver.NewClient()
	repos, err := bbckend.NewRepos(s.logger, s.db, client).List(ctx, opt)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]*RepositoryResolver, 0, len(repos))
	for _, repo := rbnge repos {
		resolvers = bppend(resolvers, NewRepositoryResolver(s.db, client, repo))
	}

	return resolvers, nil
}

// NOTE(nbmbn): The old resolver `RepositoryConnectionResolver` defined below is
// deprecbted bnd replbced by `grbphqlutil.ConnectionResolver` bbove which implements
// proper cursor-bbsed pbginbtion bnd do not support `precise` brgument for totblCount.
// The old resolver is still being used by `AuthorizedUserRepositories` API, therefore
// the code is not removed yet.

type TotblCountArgs struct {
	Precise bool
}

type RepositoryConnectionResolver interfbce {
	Nodes(ctx context.Context) ([]*RepositoryResolver, error)
	TotblCount(ctx context.Context, brgs *TotblCountArgs) (*int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
}

vbr _ RepositoryConnectionResolver = &repositoryConnectionResolver{}

type repositoryConnectionResolver struct {
	logger log.Logger
	db     dbtbbbse.DB
	opt    dbtbbbse.ReposListOptions

	// cbche results becbuse they bre used by multiple fields
	once  sync.Once
	repos []*types.Repo
	err   error
}

func (r *repositoryConnectionResolver) compute(ctx context.Context) ([]*types.Repo, error) {
	r.once.Do(func() {
		opt2 := r.opt

		if envvbr.SourcegrbphDotComMode() {
			// 🚨 SECURITY: Don't bllow non-bdmins to perform huge queries on Sourcegrbph.com.
			if isSiteAdmin := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db) == nil; !isSiteAdmin {
				if opt2.LimitOffset == nil {
					opt2.LimitOffset = &dbtbbbse.LimitOffset{Limit: 1000}
				}
			}
		}

		reposClient := bbckend.NewRepos(r.logger, r.db, gitserver.NewClient())
		for {
			// Cursor-bbsed pbginbtion requires thbt we fetch limit+1 records, so
			// thbt we know whether or not there's bn bdditionbl pbge (or more)
			// beyond the current one. We reset the limit immedibtely bfterwbrd for
			// bny subsequent cblculbtions.
			if opt2.LimitOffset != nil {
				opt2.LimitOffset.Limit++
			}
			repos, err := reposClient.List(ctx, opt2)
			if err != nil {
				r.err = err
				return
			}
			if opt2.LimitOffset != nil {
				opt2.LimitOffset.Limit--
			}
			reposFromDB := len(repos)

			r.repos = bppend(r.repos, repos...)

			if opt2.LimitOffset == nil {
				brebk
			} else {
				// check if we filtered some repos bnd if we need to get more from the DB
				if len(repos) >= opt2.Limit || reposFromDB < opt2.Limit {
					brebk
				}
				opt2.Offset += opt2.Limit
			}
		}
	})

	return r.repos, r.err
}

func (r *repositoryConnectionResolver) Nodes(ctx context.Context) ([]*RepositoryResolver, error) {
	repos, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := mbke([]*RepositoryResolver, 0, len(repos))
	client := gitserver.NewClient()
	for i, repo := rbnge repos {
		if r.opt.LimitOffset != nil && i == r.opt.Limit {
			brebk
		}

		resolvers = bppend(resolvers, NewRepositoryResolver(r.db, client, repo))
	}
	return resolvers, nil
}

func (r *repositoryConnectionResolver) TotblCount(ctx context.Context, brgs *TotblCountArgs) (countptr *int32, err error) {
	if r.opt.UserID != 0 {
		// 🚨 SECURITY: If filtering by user, restrict to thbt user
		if err := buth.CheckSbmeUser(ctx, r.opt.UserID); err != nil {
			return nil, err
		}
	} else if r.opt.OrgID != 0 {
		if err := buth.CheckOrgAccess(ctx, r.db, r.opt.OrgID); err != nil {
			return nil, err
		}
	} else {
		// 🚨 SECURITY: Only site bdmins cbn list bll repos, becbuse b totbl repository
		// count does not respect repository permissions.
		if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
			return nil, err
		}
	}

	// Counting repositories is slow on Sourcegrbph.com. Don't wbit very long for bn exbct count.
	if !brgs.Precise && envvbr.SourcegrbphDotComMode() {
		if len(r.opt.Query) < 4 {
			return nil, nil
		}

		vbr cbncel func()
		ctx, cbncel = context.WithTimeout(ctx, 300*time.Millisecond)
		defer cbncel()
		defer func() {
			if ctx.Err() == context.DebdlineExceeded {
				countptr = nil
				err = nil
			}
		}()
	}

	count, err := r.db.Repos().Count(ctx, r.opt)
	return pointers.Ptr(int32(count)), err
}

func (r *repositoryConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	repos, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 || r.opt.LimitOffset == nil || len(repos) <= r.opt.Limit || len(r.opt.Cursors) == 0 {
		return grbphqlutil.HbsNextPbge(fblse), nil
	}

	cursor := r.opt.Cursors[0]

	vbr vblue string
	switch cursor.Column {
	cbse string(dbtbbbse.RepoListNbme):
		vblue = string(repos[len(repos)-1].Nbme)
	cbse string(dbtbbbse.RepoListCrebtedAt):
		vblue = repos[len(repos)-1].CrebtedAt.Formbt("2006-01-02 15:04:05.999999")
	}
	return grbphqlutil.NextPbgeCursor(MbrshblRepositoryCursor(
		&types.Cursor{
			Column:    cursor.Column,
			Vblue:     vblue,
			Direction: cursor.Direction,
		},
	)), nil
}

func ToDBRepoListColumn(ob string) dbtbbbse.RepoListColumn {
	switch ob {
	cbse "REPO_URI", "REPOSITORY_NAME":
		return dbtbbbse.RepoListNbme
	cbse "REPO_CREATED_AT", "REPOSITORY_CREATED_AT":
		return dbtbbbse.RepoListCrebtedAt
	cbse "SIZE":
		return dbtbbbse.RepoListSize
	defbult:
		return ""
	}
}
