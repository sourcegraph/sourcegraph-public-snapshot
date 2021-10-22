package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/schema"
	"github.com/sourcegraph/sourcegraph/lib/batches/yaml"
)

func TestBatchSpecResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := actor.WithInternalActor(context.Background())
	db := dbtest.NewDB(t, "")

	cstore := store.New(db, &observation.TestContext, nil)
	repoStore := database.ReposWith(cstore)
	esStore := database.ExternalServicesWith(cstore)

	repo := newGitHubTestRepo("github.com/sourcegraph/batch-spec-test", newGitHubExternalService(t, esStore))
	if err := repoStore.Create(ctx, repo); err != nil {
		t.Fatal(err)
	}
	repoID := graphqlbackend.MarshalRepositoryID(repo.ID)

	orgname := "test-org"
	userID := ct.CreateTestUser(t, db, false).ID
	adminID := ct.CreateTestUser(t, db, true).ID
	orgID := ct.InsertTestOrg(t, db, orgname)

	spec, err := btypes.NewBatchSpecFromRaw(ct.TestRawBatchSpec)
	if err != nil {
		t.Fatal(err)
	}
	spec.UserID = userID
	spec.NamespaceOrgID = orgID
	if err := cstore.CreateBatchSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	changesetSpec, err := btypes.NewChangesetSpecFromRaw(ct.NewRawChangesetSpecGitBranch(repoID, "deadb33f"))
	if err != nil {
		t.Fatal(err)
	}
	changesetSpec.BatchSpecID = spec.ID
	changesetSpec.UserID = userID
	changesetSpec.RepoID = repo.ID

	if err := cstore.CreateChangesetSpec(ctx, changesetSpec); err != nil {
		t.Fatal(err)
	}

	matchingBatchChange := &btypes.BatchChange{
		Name:             spec.Spec.Name,
		NamespaceOrgID:   orgID,
		InitialApplierID: userID,
		LastApplierID:    userID,
		LastAppliedAt:    time.Now(),
		BatchSpecID:      spec.ID,
	}
	if err := cstore.CreateBatchChange(ctx, matchingBatchChange); err != nil {
		t.Fatal(err)
	}

	s, err := graphqlbackend.NewSchema(db, &Resolver{store: cstore}, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	apiID := string(marshalBatchSpecRandID(spec.RandID))
	userAPIID := string(graphqlbackend.MarshalUserID(userID))
	orgAPIID := string(graphqlbackend.MarshalOrgID(orgID))

	var unmarshaled interface{}
	err = json.Unmarshal([]byte(spec.RawSpec), &unmarshaled)
	if err != nil {
		t.Fatal(err)
	}

	applyUrl := fmt.Sprintf("/organizations/%s/batch-changes/apply/%s", orgname, apiID)
	want := apitest.BatchSpec{
		Typename: "BatchSpec",
		ID:       apiID,

		OriginalInput: spec.RawSpec,
		ParsedInput:   graphqlbackend.JSONValue{Value: unmarshaled},

		ApplyURL:            &applyUrl,
		Namespace:           apitest.UserOrg{ID: orgAPIID, Name: orgname},
		Creator:             &apitest.User{ID: userAPIID, DatabaseID: userID},
		ViewerCanAdminister: true,

		CreatedAt: graphqlbackend.DateTime{Time: spec.CreatedAt.Truncate(time.Second)},
		ExpiresAt: &graphqlbackend.DateTime{Time: spec.ExpiresAt().Truncate(time.Second)},

		ChangesetSpecs: apitest.ChangesetSpecConnection{
			TotalCount: 1,
			Nodes: []apitest.ChangesetSpec{
				{
					ID:       string(marshalChangesetSpecRandID(changesetSpec.RandID)),
					Typename: "VisibleChangesetSpec",
					Description: apitest.ChangesetSpecDescription{
						BaseRepository: apitest.Repository{
							ID:   string(repoID),
							Name: string(repo.Name),
						},
					},
				},
			},
		},

		DiffStat: apitest.DiffStat{
			Added:   changesetSpec.DiffStatAdded,
			Changed: changesetSpec.DiffStatChanged,
			Deleted: changesetSpec.DiffStatDeleted,
		},

		AppliesToBatchChange: apitest.BatchChange{
			ID: string(marshalBatchChangeID(matchingBatchChange.ID)),
		},

		AllCodeHosts: apitest.BatchChangesCodeHostsConnection{
			TotalCount: 1,
			Nodes:      []apitest.BatchChangesCodeHost{{ExternalServiceKind: extsvc.KindGitHub, ExternalServiceURL: "https://github.com/"}},
		},
		OnlyWithoutCredential: apitest.BatchChangesCodeHostsConnection{
			TotalCount: 1,
			Nodes:      []apitest.BatchChangesCodeHost{{ExternalServiceKind: extsvc.KindGitHub, ExternalServiceURL: "https://github.com/"}},
		},

		State: "COMPLETED",
	}

	input := map[string]interface{}{"batchSpec": apiID}
	{
		var response struct{ Node apitest.BatchSpec }
		apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryBatchSpecNode)

		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	}

	// Now create an updated changeset spec and check that we get a superseding
	// batch spec.
	sup, err := btypes.NewBatchSpecFromRaw(ct.TestRawBatchSpec)
	if err != nil {
		t.Fatal(err)
	}
	sup.UserID = userID
	sup.NamespaceOrgID = orgID
	if err := cstore.CreateBatchSpec(ctx, sup); err != nil {
		t.Fatal(err)
	}

	{
		var response struct{ Node apitest.BatchSpec }

		// Note that we have to execute as the actual user, since a superseding
		// spec isn't returned for an admin.
		apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryBatchSpecNode)

		// Expect an ID on the superseding batch spec.
		want.SupersedingBatchSpec = &apitest.BatchSpec{
			ID: string(marshalBatchSpecRandID(sup.RandID)),
		}

		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	}

	// If the superseding batch spec was created by a different user, then we
	// shouldn't return it.
	sup.UserID = adminID
	if err := cstore.UpdateBatchSpec(ctx, sup); err != nil {
		t.Fatal(err)
	}

	{
		var response struct{ Node apitest.BatchSpec }

		// Note that we have to execute as the actual user, since a superseding
		// spec isn't returned for an admin.
		apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryBatchSpecNode)

		// Expect no superseding batch spec, since this request is run as a
		// different user.
		want.SupersedingBatchSpec = nil

		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	}

	// Now soft-delete the creator and check that the batch spec is still retrievable.
	err = database.UsersWith(cstore).Delete(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	{
		var response struct{ Node apitest.BatchSpec }
		apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(adminID)), t, s, input, &response, queryBatchSpecNode)

		// Expect creator to not be returned anymore.
		want.Creator = nil
		// Expect no superseding batch spec, since this request is run as a
		// different user.
		want.SupersedingBatchSpec = nil

		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	}

	// Now hard-delete the creator and check that the batch spec is still retrievable.
	err = database.UsersWith(cstore).HardDelete(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	{
		var response struct{ Node apitest.BatchSpec }
		apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(adminID)), t, s, input, &response, queryBatchSpecNode)

		// Expect creator to not be returned anymore.
		want.Creator = nil

		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	}
}

func TestBatchSpecResolver_BatchSpecCreatedFromRaw(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := dbtest.NewDB(t, "")

	now := timeutil.Now().Truncate(time.Second)
	minAgo := func(min int) time.Time { return now.Add(time.Duration(-min) * time.Minute) }

	admin := ct.CreateTestUser(t, db, true)
	adminCtx := actor.WithActor(ctx, actor.FromUser(admin.ID))

	rs, _ := ct.CreateTestRepos(t, ctx, db, 3)

	bstore := store.New(db, &observation.TestContext, nil)

	svc := service.New(bstore)
	spec, err := svc.CreateBatchSpecFromRaw(adminCtx, service.CreateBatchSpecFromRawOpts{
		RawSpec:         ct.TestRawBatchSpecYAML,
		NamespaceUserID: admin.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	resolutionJob, err := bstore.GetBatchSpecResolutionJob(ctx, store.GetBatchSpecResolutionJobOpts{
		BatchSpecID: spec.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	s, err := graphqlbackend.NewSchema(db, &Resolver{store: bstore}, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var unmarshaled interface{}
	err = yaml.UnmarshalValidate(schema.BatchSpecJSON, []byte(spec.RawSpec), &unmarshaled)
	if err != nil {
		t.Fatal(err)
	}

	apiID := string(marshalBatchSpecRandID(spec.RandID))
	adminAPIID := string(graphqlbackend.MarshalUserID(admin.ID))

	applyUrl := fmt.Sprintf("/users/%s/batch-changes/apply/%s", admin.Username, apiID)
	codeHosts := apitest.BatchChangesCodeHostsConnection{
		TotalCount: 1,
		Nodes: []apitest.BatchChangesCodeHost{
			{ExternalServiceKind: "GITHUB", ExternalServiceURL: "https://github.com/"},
		},
	}
	want := apitest.BatchSpec{
		Typename: "BatchSpec",
		ID:       apiID,

		OriginalInput: spec.RawSpec,
		ParsedInput:   graphqlbackend.JSONValue{Value: unmarshaled},

		Namespace:           apitest.UserOrg{ID: adminAPIID, DatabaseID: admin.ID, SiteAdmin: true},
		Creator:             &apitest.User{ID: adminAPIID, DatabaseID: admin.ID, SiteAdmin: true},
		ViewerCanAdminister: true,

		AllCodeHosts:          codeHosts,
		OnlyWithoutCredential: codeHosts,

		CreatedAt: graphqlbackend.DateTime{Time: spec.CreatedAt.Truncate(time.Second)},
		ExpiresAt: &graphqlbackend.DateTime{Time: spec.ExpiresAt().Truncate(time.Second)},

		ChangesetSpecs: apitest.ChangesetSpecConnection{
			Nodes: []apitest.ChangesetSpec{},
		},

		State: "PENDING",
		WorkspaceResolution: apitest.BatchSpecWorkspaceResolution{
			State: resolutionJob.State.ToGraphQL(),
		},
	}

	queryAndAssertBatchSpec(t, adminCtx, s, apiID, want)

	// Complete the workspace resolution
	var workspaces []*btypes.BatchSpecWorkspace
	for _, repo := range rs {
		ws := &btypes.BatchSpecWorkspace{BatchSpecID: spec.ID, RepoID: repo.ID, Steps: []batcheslib.Step{}}
		if err := bstore.CreateBatchSpecWorkspace(ctx, ws); err != nil {
			t.Fatal(err)
		}
		workspaces = append(workspaces, ws)
	}

	setResolutionJobState(t, ctx, bstore, resolutionJob, btypes.BatchSpecResolutionJobStateCompleted)
	want.WorkspaceResolution.State = btypes.BatchSpecResolutionJobStateCompleted.ToGraphQL()
	queryAndAssertBatchSpec(t, adminCtx, s, apiID, want)

	// Now enqueue jobs
	var jobs []*btypes.BatchSpecWorkspaceExecutionJob
	for _, ws := range workspaces {
		job := &btypes.BatchSpecWorkspaceExecutionJob{BatchSpecWorkspaceID: ws.ID}
		if err := ct.CreateBatchSpecWorkspaceExecutionJob(ctx, bstore, store.ScanBatchSpecWorkspaceExecutionJob, job); err != nil {
			t.Fatal(err)
		}
		jobs = append(jobs, job)
	}

	want.State = "QUEUED"
	queryAndAssertBatchSpec(t, adminCtx, s, apiID, want)

	// 1/3 jobs processing
	jobs[1].StartedAt = minAgo(99)
	setJobProcessing(t, ctx, bstore, jobs[1])
	want.State = "PROCESSING"
	want.StartedAt = graphqlbackend.DateTime{Time: jobs[1].StartedAt}
	queryAndAssertBatchSpec(t, adminCtx, s, apiID, want)

	// 3/3 processing
	setJobProcessing(t, ctx, bstore, jobs[0])
	setJobProcessing(t, ctx, bstore, jobs[2])
	// Expect same state
	queryAndAssertBatchSpec(t, adminCtx, s, apiID, want)

	// 1/3 jobs complete, 2/3 processing
	jobs[2].FinishedAt = minAgo(30)
	setJobCompleted(t, ctx, bstore, jobs[2])
	// Expect same state
	queryAndAssertBatchSpec(t, adminCtx, s, apiID, want)

	// 3/3 jobs complete
	jobs[0].FinishedAt = minAgo(9)
	jobs[1].FinishedAt = minAgo(15)
	setJobCompleted(t, ctx, bstore, jobs[0])
	setJobCompleted(t, ctx, bstore, jobs[1])
	want.State = "COMPLETED"
	want.ApplyURL = &applyUrl
	want.FinishedAt = graphqlbackend.DateTime{Time: jobs[0].FinishedAt}
	queryAndAssertBatchSpec(t, adminCtx, s, apiID, want)
	want.ApplyURL = nil

	// 1/3 jobs is failed, 2/3 completed
	setJobFailed(t, ctx, bstore, jobs[1])
	want.State = "FAILED"
	queryAndAssertBatchSpec(t, adminCtx, s, apiID, want)

	// 1/3 jobs is failed, 2/3 still processing
	setJobProcessing(t, ctx, bstore, jobs[0])
	setJobProcessing(t, ctx, bstore, jobs[2])
	want.State = "PROCESSING"
	want.FinishedAt = graphqlbackend.DateTime{}
	queryAndAssertBatchSpec(t, adminCtx, s, apiID, want)

	// 3/3 jobs canceling and processing
	setJobCanceling(t, ctx, bstore, jobs[0])
	setJobCanceling(t, ctx, bstore, jobs[1])
	setJobCanceling(t, ctx, bstore, jobs[2])

	want.State = "CANCELING"
	queryAndAssertBatchSpec(t, adminCtx, s, apiID, want)

	// 3/3 canceling and failed
	jobs[0].FinishedAt = minAgo(9)
	jobs[1].FinishedAt = minAgo(15)
	jobs[2].FinishedAt = minAgo(30)
	setJobCanceled(t, ctx, bstore, jobs[0])
	setJobCanceled(t, ctx, bstore, jobs[1])
	setJobCanceled(t, ctx, bstore, jobs[2])

	want.State = "CANCELED"
	want.FinishedAt = graphqlbackend.DateTime{Time: jobs[0].FinishedAt}
	queryAndAssertBatchSpec(t, adminCtx, s, apiID, want)
}

func queryAndAssertBatchSpec(t *testing.T, ctx context.Context, s *graphql.Schema, id string, want apitest.BatchSpec) {
	t.Helper()

	input := map[string]interface{}{"batchSpec": id}

	var response struct{ Node apitest.BatchSpec }

	apitest.MustExec(ctx, t, s, input, &response, queryBatchSpecNode)

	if diff := cmp.Diff(want, response.Node); diff != "" {
		t.Fatalf("unexpected batch spec (-want +got):\n%s", diff)
	}
}

func setJobProcessing(t *testing.T, ctx context.Context, s *store.Store, job *btypes.BatchSpecWorkspaceExecutionJob) {
	t.Helper()
	job.State = btypes.BatchSpecWorkspaceExecutionJobStateProcessing
	if job.StartedAt.IsZero() {
		job.StartedAt = time.Now().Add(-5 * time.Minute)
	}
	job.FinishedAt = time.Time{}
	job.Cancel = false
	ct.UpdateJobState(t, ctx, s, job)
}

func setJobCompleted(t *testing.T, ctx context.Context, s *store.Store, job *btypes.BatchSpecWorkspaceExecutionJob) {
	t.Helper()
	job.State = btypes.BatchSpecWorkspaceExecutionJobStateCompleted
	if job.StartedAt.IsZero() {
		job.StartedAt = time.Now().Add(-5 * time.Minute)
	}
	if job.FinishedAt.IsZero() {
		job.FinishedAt = time.Now()
	}
	job.Cancel = false
	ct.UpdateJobState(t, ctx, s, job)
}

func setJobFailed(t *testing.T, ctx context.Context, s *store.Store, job *btypes.BatchSpecWorkspaceExecutionJob) {
	t.Helper()
	job.State = btypes.BatchSpecWorkspaceExecutionJobStateFailed
	if job.StartedAt.IsZero() {
		job.StartedAt = time.Now().Add(-5 * time.Minute)
	}
	if job.FinishedAt.IsZero() {
		job.FinishedAt = time.Now()
	}
	job.Cancel = false
	ct.UpdateJobState(t, ctx, s, job)
}

func setJobCanceling(t *testing.T, ctx context.Context, s *store.Store, job *btypes.BatchSpecWorkspaceExecutionJob) {
	t.Helper()
	job.State = btypes.BatchSpecWorkspaceExecutionJobStateProcessing
	if job.StartedAt.IsZero() {
		job.StartedAt = time.Now().Add(-5 * time.Minute)
	}
	job.FinishedAt = time.Time{}
	job.Cancel = true
	ct.UpdateJobState(t, ctx, s, job)
}

func setJobCanceled(t *testing.T, ctx context.Context, s *store.Store, job *btypes.BatchSpecWorkspaceExecutionJob) {
	t.Helper()
	job.State = btypes.BatchSpecWorkspaceExecutionJobStateFailed
	if job.StartedAt.IsZero() {
		job.StartedAt = time.Now().Add(-5 * time.Minute)
	}
	if job.FinishedAt.IsZero() {
		job.FinishedAt = time.Now()
	}
	job.Cancel = true
	ct.UpdateJobState(t, ctx, s, job)
}

func setResolutionJobState(t *testing.T, ctx context.Context, s *store.Store, job *btypes.BatchSpecResolutionJob, state btypes.BatchSpecResolutionJobState) {
	t.Helper()

	job.State = state

	err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_resolution_jobs SET state = %s WHERE id = %s", job.State, job.ID))
	if err != nil {
		t.Fatalf("failed to set resolution job state: %s", err)
	}
}

const queryBatchSpecNode = `
fragment u on User { id, databaseID, siteAdmin }
fragment o on Org  { id, name }

query($batchSpec: ID!) {
  node(id: $batchSpec) {
    __typename

    ... on BatchSpec {
      id
      originalInput
      parsedInput

      creator  { ...u }
      namespace {
        ... on User { ...u }
        ... on Org  { ...o }
      }

      applyURL
      viewerCanAdminister

      createdAt
      expiresAt

      diffStat { added, deleted, changed }

	  appliesToBatchChange { id }
	  supersedingBatchSpec { id }

	  allCodeHosts: viewerBatchChangesCodeHosts {
		totalCount
		  nodes {
			  externalServiceKind
			  externalServiceURL
		  }
	  }

	  onlyWithoutCredential: viewerBatchChangesCodeHosts(onlyWithoutCredential: true) {
		  totalCount
		  nodes {
			  externalServiceKind
			  externalServiceURL
		  }
	  }

      changesetSpecs(first: 100) {
        totalCount

        nodes {
          __typename
          type

          ... on HiddenChangesetSpec {
            id
          }

          ... on VisibleChangesetSpec {
            id

            description {
              ... on ExistingChangesetReference {
                baseRepository {
                  id
                  name
                }
              }

              ... on GitBranchChangesetDescription {
                baseRepository {
                  id
                  name
                }
              }
            }
          }
        }
	  }

      state
      workspaceResolution {
        state
      }
      startedAt
      finishedAt
    }
  }
}
`
