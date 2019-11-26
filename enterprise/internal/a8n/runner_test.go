package a8n

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

type refAndTarget struct {
	ref    string
	target string
}

func TestRunner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}

	store := NewStoreWithClock(dbconn.Global, clock)

	defaultBranches := []refAndTarget{
		{"master", "fc21c1a0a79047416c14642b3ca964faba9442e2"},
		{"develop", "f3c08ec74a9b3f8af7b5609c9f47cfcb3dc6949b"},
		{"staging", "09d6921f5ccae24dc2cb3ca2cf263a05e547cf4f"},
	}

	rs := []*repos.Repo{
		testRepo(0, github.ServiceType),
		testRepo(1, github.ServiceType),
		testRepo(2, bitbucketserver.ServiceType),
		// Unsupported for now, filtered out
		testRepo(3, gitlab.ServiceType),
	}

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	err := reposStore.UpsertRepos(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	testPlan := &a8n.CampaignPlan{CampaignType: "test", Arguments: `{}`}

	tests := []struct {
		name string

		search       repoSearch
		commitID     repoDefaultBranch
		campaignType CampaignType

		runErr string

		wantPlan *a8n.CampaignPlan
		wantJobs func(plan *a8n.CampaignPlan, rs []*repos.Repo, branches []refAndTarget) []*a8n.CampaignJob
	}{
		{
			name: "no search results",
			search: func(ctx context.Context, query string) ([]*graphqlbackend.RepositoryResolver, error) {
				return []*graphqlbackend.RepositoryResolver{}, nil
			},
			commitID:     yieldDefaultBranches(defaultBranches),
			campaignType: &testCampaignType{},
			wantPlan: func() *a8n.CampaignPlan {
				p := testPlan.Clone()
				p.CreatedAt = now
				p.UpdatedAt = now
				return p
			}(),
			wantJobs: wantNoJobs,
		},
		{
			name: "search error",
			search: func(ctx context.Context, query string) ([]*graphqlbackend.RepositoryResolver, error) {
				return nil, errors.New("search failed")
			},
			commitID:     yieldDefaultBranches(defaultBranches),
			campaignType: &testCampaignType{},
			runErr:       "search failed",
			wantPlan:     nil,
			wantJobs:     wantNoJobs,
		},
		{
			name: "too many search results",
			search: func(ctx context.Context, query string) ([]*graphqlbackend.RepositoryResolver, error) {
				m, _ := strconv.ParseInt(maxRepositories, 10, 64)
				count := m + 1

				resolvers := make([]*graphqlbackend.RepositoryResolver, count)
				for i := 0; i < int(count); i++ {
					resolvers[i] = repoToResolver(rs[i%len(rs)])
				}
				return resolvers, nil
			},
			commitID:     yieldDefaultBranches(defaultBranches),
			campaignType: &testCampaignType{},
			runErr:       ErrTooManyResults.Error(),
			wantPlan:     nil,
			wantJobs:     wantNoJobs,
		},
		{
			name:         "multi search results and successfull execution",
			search:       yieldRepos(rs...),
			commitID:     yieldDefaultBranches(defaultBranches),
			campaignType: &testCampaignType{diff: testDiff},
			wantPlan: func() *a8n.CampaignPlan {
				p := testPlan.Clone()
				p.CreatedAt = now
				p.UpdatedAt = now
				return p
			}(),
			wantJobs: func(plan *a8n.CampaignPlan, rs []*repos.Repo, branches []refAndTarget) []*a8n.CampaignJob {
				return []*a8n.CampaignJob{
					{
						CampaignPlanID: plan.ID,
						RepoID:         int32(rs[0].ID),
						Diff:           testDiff,
						Rev:            api.CommitID(branches[0].target),
						BaseRef:        branches[0].ref,
						CreatedAt:      now,
						UpdatedAt:      now,
						StartedAt:      now,
						FinishedAt:     now,
					},
					{
						CampaignPlanID: plan.ID,
						RepoID:         int32(rs[1].ID),
						Diff:           testDiff,
						Rev:            api.CommitID(branches[1].target),
						BaseRef:        branches[1].ref,
						CreatedAt:      now,
						UpdatedAt:      now,
						StartedAt:      now,
						FinishedAt:     now,
					},
					{
						CampaignPlanID: plan.ID,
						RepoID:         int32(rs[2].ID),
						Diff:           testDiff,
						Rev:            api.CommitID(branches[2].target),
						BaseRef:        branches[2].ref,
						CreatedAt:      now,
						UpdatedAt:      now,
						StartedAt:      now,
						FinishedAt:     now,
					},
				}
			},
		},
		{
			name:         "multi search results but getting a commit ID fails",
			search:       yieldRepos(rs...),
			commitID:     errorOnCall(yieldDefaultBranches(defaultBranches), 2, errors.New("no commit ID found")),
			campaignType: &testCampaignType{diff: testDiff},
			runErr:       "no commit ID found",
			wantPlan:     nil,
			wantJobs:     wantNoJobs,
		},
		{
			name:         "two search results but one repo has no default branch",
			search:       yieldRepos(rs[0], rs[1]),
			commitID:     errorOnCall(yieldDefaultBranches(defaultBranches), 1, ErrNoDefaultBranch),
			campaignType: &testCampaignType{diff: testDiff},
			wantPlan: func() *a8n.CampaignPlan {
				p := testPlan.Clone()
				p.CreatedAt = now
				p.UpdatedAt = now
				return p
			}(),
			wantJobs: func(plan *a8n.CampaignPlan, rs []*repos.Repo, branches []refAndTarget) []*a8n.CampaignJob {
				return []*a8n.CampaignJob{
					{
						CampaignPlanID: plan.ID,
						RepoID:         int32(rs[0].ID),
						Diff:           testDiff,
						Rev:            api.CommitID(branches[0].target),
						BaseRef:        branches[0].ref,
						CreatedAt:      now,
						UpdatedAt:      now,
						StartedAt:      now,
						FinishedAt:     now,
					},
				}
			},
		},
		{
			name:     "generating diff fails",
			search:   yieldRepos(rs[0]),
			commitID: yieldDefaultBranches(defaultBranches),
			campaignType: &testCampaignType{
				diff:    testDiff,
				diffErr: "could not generate diff",
			},
			wantPlan: func() *a8n.CampaignPlan {
				p := testPlan.Clone()
				p.CreatedAt = now
				p.UpdatedAt = now
				return p
			}(),
			wantJobs: func(plan *a8n.CampaignPlan, rs []*repos.Repo, branches []refAndTarget) []*a8n.CampaignJob {
				return []*a8n.CampaignJob{
					{
						CampaignPlanID: plan.ID,
						RepoID:         int32(rs[0].ID),
						Diff:           "",
						Error:          "could not generate diff",
						Rev:            api.CommitID(branches[0].target),
						BaseRef:        branches[0].ref,
						CreatedAt:      now,
						UpdatedAt:      now,
						StartedAt:      now,
						FinishedAt:     now,
					},
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.runErr == "" {
				tc.runErr = "<nil>"
			}

			plan := testPlan.Clone()

			runner := NewRunnerWithClock(store, tc.campaignType, tc.search, tc.commitID, clock)
			err := runner.Run(ctx, plan)
			if have, want := fmt.Sprint(err), tc.runErr; have != want {
				t.Fatalf("have runner.Run error: %q\nwant error: %q", have, want)
			}

			waitRunner(t, runner)

			if tc.wantPlan == nil && plan.ID == 0 {
				return
			}

			havePlan, err := store.GetCampaignPlan(ctx, GetCampaignPlanOpts{ID: plan.ID})
			if err == ErrNoResults && tc.wantPlan == nil {
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			planIgnore := cmpopts.IgnoreFields(a8n.CampaignPlan{}, "ID")
			if diff := cmp.Diff(havePlan, tc.wantPlan, planIgnore); diff != "" {
				t.Fatalf("CampaignPlan diff: %s", diff)
			}

			haveJobs, _, err := store.ListCampaignJobs(ctx, ListCampaignJobsOpts{
				CampaignPlanID: plan.ID,
			})
			if err != nil {
				t.Fatal(err)
			}

			sort.Slice(haveJobs, func(i, j int) bool {
				return haveJobs[i].RepoID < haveJobs[j].RepoID
			})

			wantJobs := tc.wantJobs(plan, rs, defaultBranches)
			jobIgnore := cmpopts.IgnoreFields(a8n.CampaignJob{}, "ID")
			if diff := cmp.Diff(haveJobs, wantJobs, jobIgnore); diff != "" {
				t.Fatalf("CampaignJobs diff: %s", diff)
			}
		})
	}

}

type testCampaignType struct {
	diff    string
	diffErr string
}

func (t *testCampaignType) searchQuery() string { return "" }
func (t *testCampaignType) generateDiff(ctx context.Context, repo api.RepoName, commit api.CommitID) (string, error) {
	if t.diffErr != "" {
		return "", errors.New(t.diffErr)
	}
	return t.diff, nil
}

const testDiff = `diff --git a/README.md b/README.md
index 851b23a..140f333 100644
--- a/README.md
+++ b/README.md
@@ -1,3 +1,4 @@
 # README
 
+Let's add a line here.
 This file is hostEd at sourcegraph.com and is a test file.
`

func waitRunner(t *testing.T, r *Runner) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		defer func() { close(done) }()
		err := r.Wait()
		if err != nil {
			t.Errorf("runner.Wait failed: %s", err)
		}
	}()

	select {
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout reached")
	case <-done:
	}
}

func wantNoJobs(plan *a8n.CampaignPlan, rs []*repos.Repo, branches []refAndTarget) []*a8n.CampaignJob {
	return []*a8n.CampaignJob{}
}

func testRepo(num int, serviceType string) *repos.Repo {
	extSvcID := fmt.Sprintf("extsvc:%s:%d", serviceType, num)

	return &repos.Repo{
		Name:    fmt.Sprintf("repo-%d", num),
		URI:     fmt.Sprintf("repo-%d", num),
		Enabled: true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          fmt.Sprintf("external-id-%d", num),
			ServiceType: serviceType,
			ServiceID:   "https://example.com/",
		},
		Sources: map[string]*repos.SourceInfo{
			extSvcID: {
				ID:       extSvcID,
				CloneURL: "https://secrettoken@github.com/sourcegraph/sourcegraph",
			},
		},
	}
}

func repoToResolver(r *repos.Repo) *graphqlbackend.RepositoryResolver {
	return graphqlbackend.NewRepositoryResolver(&types.Repo{
		ID:           api.RepoID(r.ID),
		ExternalRepo: r.ExternalRepo,
		Name:         api.RepoName(r.Name),
		RepoFields: &types.RepoFields{
			URI:         r.URI,
			Description: r.Description,
			Language:    r.Language,
			Fork:        r.Fork,
		},
	})
}

func yieldDefaultBranches(branches []refAndTarget) repoDefaultBranch {
	count := 0
	return func(ctx context.Context, repo *graphqlbackend.RepositoryResolver) (string, api.CommitID, error) {
		branch := "invalid"
		id := api.CommitID("invalid")

		if count >= len(branches) {
			return branch, id, errors.New("exhausted commit ids")
		}

		branch = branches[count].ref
		id = api.CommitID(branches[count].target)

		count++

		return branch, id, nil
	}
}

func errorOnCall(f repoDefaultBranch, num int, err error) repoDefaultBranch {
	count := 0

	return func(ctx context.Context, repo *graphqlbackend.RepositoryResolver) (string, api.CommitID, error) {
		branch := "invalid"
		id := api.CommitID("invalid")

		if count == num {
			return branch, id, err
		}

		count++

		return f(ctx, repo)
	}
}

func yieldRepos(rs ...*repos.Repo) repoSearch {
	resolvers := make([]*graphqlbackend.RepositoryResolver, len(rs))
	for i, r := range rs {
		resolvers[i] = repoToResolver(r)
	}
	return func(ctx context.Context, query string) ([]*graphqlbackend.RepositoryResolver, error) {
		return resolvers, nil
	}
}
