package graphqlbackend

import (
	"context"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var (
	cf = httpcli.NewExternalHTTPClientFactory()
)

type codeHostRepositoryConnectionResolver struct {
	userID   int32
	codeHost int64
	query    string

	once  sync.Once
	nodes []*codeHostRepositoryResolver
	err   error
	db    dbutil.DB
}

func (r *codeHostRepositoryConnectionResolver) Nodes(ctx context.Context) ([]*codeHostRepositoryResolver, error) {
	r.once.Do(func() {
		var (
			svcs []*types.ExternalService
			err  error
		)
		// get all external services for user, or for the specified external service
		if r.codeHost == 0 {
			svcs, err = database.GlobalExternalServices.List(ctx, database.ExternalServicesListOptions{NamespaceUserID: r.userID})
			if err != nil {
				r.err = err
				return
			}
		} else {
			svc, err := database.GlobalExternalServices.GetByID(ctx, r.codeHost)
			if err != nil {
				r.err = err
				return
			}
			// 🚨 SECURITY: if the user doesn't own this service, check they're site admin
			if err := backend.CheckUserIsSiteAdmin(ctx, r.userID); svc.NamespaceUserID != r.userID && err != nil {
				r.err = err
				return
			}
			svcs = []*types.ExternalService{svc}
		}
		// get Source for all external services
		var (
			results  = make(chan []types.CodeHostRepository)
			g, ctx   = errgroup.WithContext(ctx)
			svcsByID = make(map[int64]*types.ExternalService)
		)
		for _, svc := range svcs {
			svcsByID[svc.ID] = svc
			src, err := repos.NewSource(svc, cf)
			if err != nil {
				r.err = err
				return
			}
			if af, ok := src.(repos.AffiliatedRepositorySource); ok {
				g.Go(func() error {
					repos, err := af.AffiliatedRepositories(ctx)
					if err != nil {
						return err
					}
					select {
					case results <- repos:
					case <-ctx.Done():
						return ctx.Err()
					}
					return nil
				})
			}
		}
		go func() {
			// wait for all sources to return their repos
			err = g.Wait()
			// signal the collector to finish
			close(results)
		}()

		// are we allowed to show the user private repos?
		allowPrivate, err := allowPrivate(ctx, r.db, r.userID)
		if err != nil {
			r.err = err
			return
		}

		// collect all results
		r.nodes = []*codeHostRepositoryResolver{}
		for repos := range results {
			for _, repo := range repos {
				repo := repo
				if r.query != "" && !strings.Contains(strings.ToLower(repo.Name), r.query) {
					continue
				}
				if !allowPrivate && repo.Private {
					continue
				}
				r.nodes = append(r.nodes, &codeHostRepositoryResolver{
					db:       r.db,
					codeHost: svcsByID[repo.CodeHostID],
					repo:     &repo,
				})
			}
		}
		sort.Slice(r.nodes, func(i, j int) bool {
			return r.nodes[i].repo.Name < r.nodes[j].repo.Name
		})
	})
	return r.nodes, r.err
}

type codeHostRepositoryResolver struct {
	repo     *types.CodeHostRepository
	codeHost *types.ExternalService
	db       dbutil.DB
}

func (r *codeHostRepositoryResolver) Name() string {
	return r.repo.Name
}

func (r *codeHostRepositoryResolver) Private() bool {
	return r.repo.Private
}

func (r *codeHostRepositoryResolver) CodeHost(ctx context.Context) *externalServiceResolver {
	return &externalServiceResolver{
		db:              r.db,
		externalService: r.codeHost,
	}
}

func allowPrivate(ctx context.Context, db dbutil.DB, userID int32) (bool, error) {
	mode, err := database.Users(db).UserAllowedExternalServices(ctx, userID)
	if err != nil {
		return false, err
	}
	return mode == conf.ExternalServiceModeAll, nil
}
