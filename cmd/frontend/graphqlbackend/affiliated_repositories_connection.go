package graphqlbackend

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	cf = httpcli.ExternalClientFactory
)

type codeHostRepositoryResolver struct {
	repo         *types.CodeHostRepository
	codeHost     *types.ExternalService
	db           database.DB
	codeHostErrs string
}

type svcResult struct {
	nodes []*codeHostRepositoryResolver
	err   string
}

type affiliatedRepositoriesConnection struct {
	userID     int32
	orgID      int32
	codeHost   int64
	query      string
	once       sync.Once
	svcResults []*svcResult
	// nodes    []*codeHostRepositoryResolver
	err error
	db  database.DB
}

func (a *affiliatedRepositoriesConnection) SvcResults(ctx context.Context) ([]*svcResult, error) {
	a.once.Do(func() {
		var (
			svcs []*types.ExternalService
			err  error
		)
		// get all external services for the user, the organization, or for the specified external service
		if a.codeHost == 0 {
			svcs, err = a.db.ExternalServices().List(ctx, database.ExternalServicesListOptions{
				NamespaceUserID: a.userID,
				NamespaceOrgID:  a.orgID,
			})
			if err != nil {
				a.err = err
				return
			}

		} else {
			svc, err := a.db.ExternalServices().GetByID(ctx, a.codeHost)
			if err != nil {
				a.err = err
				return
			}
			// 🚨 SECURITY: check if user can access external service
			err = backend.CheckExternalServiceAccess(ctx, a.db, svc.NamespaceUserID, svc.NamespaceOrgID)
			if err != nil {
				a.err = err
				return
			}
			svcs = append(svcs, svc)
		}

		type affiliatedResult struct {
			svcID int64
			repos []types.CodeHostRepository
			err   error
		}

		// get Source for all external services
		var (
			results  = make(chan affiliatedResult, len(svcs))
			svcsByID = make(map[int64]*types.ExternalService)
			pending  int
		)
		for i, svc := range svcs {
			svcsByID[svc.ID] = svc
			src, err := repos.NewSource(a.db.ExternalServices(), svc, cf)
			if err != nil {
				a.err = err
				return
			}
			af, ok := src.(repos.AffiliatedRepositorySource)
			if !ok {
				continue
			}
			pending++

			svcID := svc.ID
			goroutine.Go(func() {
				affiliated, err := af.AffiliatedRepositories(ctx)
				results <- affiliatedResult{
					svcID: svcID,
					repos: affiliated,
					err:   err,
				}
			})
		}

		// are we allowed to show the private repos?
		allowPrivate, err := allowPrivate(ctx, a.db, a.userID, a.orgID)
		if err != nil {
			a.err = err
			return
		}

		var fetchErrors []error
		var nodes []*codeHostRepositoryResolver
		a.svcResults = []*svcResult{}

		for i := 0; i < pending; i++ {
			select {
			case result := <-results:
				if result.err != nil { // if an errors comes from results.err, repos is an empty array
					// An error from one code is not fatal
					log15.Error("getting affiliated repos", "externalServiceId", result.svcID, "err", result.err)
					fetchErrors = append(fetchErrors, result.err)

					// Get the error from the code host so we can return it.
					errMessage := "Error fetching repos from " + svcsByID[result.svcID].DisplayName + ": " + result.err.Error()
					a.svcResults = append(a.svcResults, &svcResult{
						nodes: nn,
						err:   errMessage,
					})

					continue
				}

				for _, repo := range result.repos {
					if a.query != "" && !strings.Contains(strings.ToLower(repo.Name), a.query) {
						continue
					}
					if !allowPrivate && repo.Private {
						continue
					}

					repo := repo
					nodes = append(nodes, &codeHostRepositoryResolver{
						repo:     &repo,
						codeHost: svcsByID[repo.CodeHostID],
						db:       a.db,
					})
				}

				var errMessage string
				a.svcResults = append(a.svcResults, &svcResult{
					nodes: nodes,
					err:   errMessage,
				})

			case <-ctx.Done():
				a.err = ctx.Err()
				return
			}
		}

		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].repo.Name < nodes[j].repo.Name
		})

		if len(fetchErrors) == pending {
			// All hosts failed
			for _, aa := range a.svcResults {
				aa.nodes = nil
			}
			a.err = errors.New("failed to fetch from any code host")
		}
	})

	if envvar.SourcegraphDotComMode() && a.orgID != 0 {
		for _, aa := range a.svcResults {
			a.db.OrgStats().Upsert(ctx, a.orgID, int32(len(aa.nodes)))
		}
	}

	// DEBUGGING BLOCK:
	fmt.Println("------ DEBUGGING - START")
	for i, aaa := range a.svcResults {
		fmt.Println(i, len(aaa.nodes), aaa.err)
	}
	fmt.Println("------ END")

	return a.svcResults, a.err
}

func (s *svcResult) Nodes(ctx context.Context) ([]*codeHostRepositoryResolver, error) {
	return nil, nil
}

func (s *svcResult) Err(ctx context.Context) (*string, error) {
	return nil, nil
}

func setErrorPerCodeHost(dict map[int64]string, id int64, err string) map[int64]string {
	_, ok := dict[id]
	if ok {
		return dict
	}
	dict[id] = err
	return dict
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

func (r *codeHostRepositoryResolver) CodeHostErrs() *string {
	return &r.codeHostErrs
}

func allowPrivate(ctx context.Context, db database.DB, userID, orgID int32) (bool, error) {
	if userID > 0 {
		mode, err := db.Users().UserAllowedExternalServices(ctx, userID)
		if err != nil {
			return false, err
		}
		return mode == conf.ExternalServiceModeAll, nil
	}
	if orgID > 0 {
		return true, nil
	}

	return false, errors.New("either userID or orgID expected to be defined")
}
