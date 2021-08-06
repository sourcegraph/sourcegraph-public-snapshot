package resolvers

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestChangesetApplyPreviewResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := actor.WithInternalActor(context.Background())
	db := dbtest.NewDB(t, "")

	userID := ct.CreateTestUser(t, db, false).ID

	cstore := store.New(db, &observation.TestContext, nil)

	// Create a batch spec for the target batch change.
	oldBatchSpec := &btypes.BatchSpec{
		UserID:          userID,
		NamespaceUserID: userID,
	}
	if err := cstore.CreateBatchSpec(ctx, oldBatchSpec); err != nil {
		t.Fatal(err)
	}
	// Create a batch change and create a new spec targetting the same batch change again.
	batchChangeName := "test-apply-preview-resolver"
	batchChange := ct.CreateBatchChange(t, ctx, cstore, batchChangeName, userID, oldBatchSpec.ID)
	batchSpec := ct.CreateBatchSpec(t, ctx, cstore, batchChangeName, userID)

	esStore := database.ExternalServicesWith(cstore)
	repoStore := database.ReposWith(cstore)

	rs := make([]*types.Repo, 0, 3)
	for i := 0; i < cap(rs); i++ {
		name := fmt.Sprintf("github.com/sourcegraph/test-changeset-apply-preview-repo-%d", i)
		r := newGitHubTestRepo(name, newGitHubExternalService(t, esStore))
		if err := repoStore.Create(ctx, r); err != nil {
			t.Fatal(err)
		}
		rs = append(rs, r)
	}

	changesetSpecs := make([]*btypes.ChangesetSpec, 0, 2)
	for i, r := range rs[:2] {
		s := ct.CreateChangesetSpec(t, ctx, cstore, ct.TestSpecOpts{
			BatchSpec: batchSpec.ID,
			User:      userID,
			Repo:      r.ID,
			HeadRef:   fmt.Sprintf("d34db33f-%d", i),
		})

		changesetSpecs = append(changesetSpecs, s)
	}

	// Add one changeset that doesn't match any new spec anymore but was there before (close, detach).
	closingChangesetSpec := ct.CreateChangesetSpec(t, ctx, cstore, ct.TestSpecOpts{
		User:      userID,
		Repo:      rs[2].ID,
		BatchSpec: oldBatchSpec.ID,
		HeadRef:   "d34db33f-2",
	})
	closingChangeset := ct.CreateChangeset(t, ctx, cstore, ct.TestChangesetOpts{
		Repo:             rs[2].ID,
		BatchChange:      batchChange.ID,
		CurrentSpec:      closingChangesetSpec.ID,
		PublicationState: btypes.ChangesetPublicationStatePublished,
	})

	// Add one changeset that doesn't matches a new spec (update).
	updatedChangesetSpec := ct.CreateChangesetSpec(t, ctx, cstore, ct.TestSpecOpts{
		BatchSpec: oldBatchSpec.ID,
		User:      userID,
		Repo:      changesetSpecs[1].RepoID,
		HeadRef:   changesetSpecs[1].Spec.HeadRef,
	})
	updatedChangeset := ct.CreateChangeset(t, ctx, cstore, ct.TestChangesetOpts{
		Repo:               rs[1].ID,
		BatchChange:        batchChange.ID,
		CurrentSpec:        updatedChangesetSpec.ID,
		PublicationState:   btypes.ChangesetPublicationStatePublished,
		OwnedByBatchChange: batchChange.ID,
	})

	s, err := graphqlbackend.NewSchema(db, &Resolver{store: cstore}, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	apiID := string(marshalBatchSpecRandID(batchSpec.RandID))

	input := map[string]interface{}{"batchSpec": apiID}
	var response struct{ Node apitest.BatchSpec }
	apitest.MustExec(ctx, t, s, input, &response, queryChangesetApplyPreview)

	haveApplyPreview := response.Node.ApplyPreview.Nodes

	wantApplyPreview := []apitest.ChangesetApplyPreview{
		{
			Typename:   "VisibleChangesetApplyPreview",
			Operations: []btypes.ReconcilerOperation{btypes.ReconcilerOperationDetach},
			Targets: apitest.ChangesetApplyPreviewTargets{
				Typename:  "VisibleApplyPreviewTargetsDetach",
				Changeset: apitest.Changeset{ID: string(marshalChangesetID(closingChangeset.ID))},
			},
		},
		{
			Typename:   "VisibleChangesetApplyPreview",
			Operations: []btypes.ReconcilerOperation{},
			Targets: apitest.ChangesetApplyPreviewTargets{
				Typename:      "VisibleApplyPreviewTargetsAttach",
				ChangesetSpec: apitest.ChangesetSpec{ID: string(marshalChangesetSpecRandID(changesetSpecs[0].RandID))},
			},
		},
		{
			Typename:   "VisibleChangesetApplyPreview",
			Operations: []btypes.ReconcilerOperation{},
			Targets: apitest.ChangesetApplyPreviewTargets{
				Typename:      "VisibleApplyPreviewTargetsUpdate",
				ChangesetSpec: apitest.ChangesetSpec{ID: string(marshalChangesetSpecRandID(changesetSpecs[1].RandID))},
				Changeset:     apitest.Changeset{ID: string(marshalChangesetID(updatedChangeset.ID))},
			},
		},
	}

	if diff := cmp.Diff(wantApplyPreview, haveApplyPreview); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
	}
}

const queryChangesetApplyPreview = `
query ($batchSpec: ID!, $first: Int = 50, $after: String, $publicationStates: [ChangesetSpecPublicationStateInput!]) {
    node(id: $batchSpec) {
      __typename
      ... on BatchSpec {
        id
        applyPreview(first: $first, after: $after, publicationStates: $publicationStates) {
          totalCount
          pageInfo {
            hasNextPage
            endCursor
          }
          nodes {
            __typename
            ... on VisibleChangesetApplyPreview {
			  operations
              delta {
                titleChanged
                bodyChanged
                undraft
                baseRefChanged
                diffChanged
                commitMessageChanged
                authorNameChanged
                authorEmailChanged
              }
              targets {
                __typename
                ... on VisibleApplyPreviewTargetsAttach {
                  changesetSpec {
                    id
                  }
                }
                ... on VisibleApplyPreviewTargetsUpdate {
                  changesetSpec {
                    id
                  }
                  changeset {
                    id
                  }
                }
                ... on VisibleApplyPreviewTargetsDetach {
                  changeset {
                    id
                  }
                }
              }
            }
            ... on HiddenChangesetApplyPreview {
              operations
              targets {
                __typename
                ... on HiddenApplyPreviewTargetsAttach {
                  changesetSpec {
                    id
                  }
                }
                ... on HiddenApplyPreviewTargetsUpdate {
                  changesetSpec {
                    id
                  }
                  changeset {
                    id
                  }
                }
                ... on HiddenApplyPreviewTargetsDetach {
                  changeset {
                    id
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`

func TestChangesetApplyPreviewResolverWithPublicationStates(t *testing.T) {
	// We have multiple scenarios to test here: these essentially act as
	// integration tests for the applyPreview() resolver when publication states
	// are set.
	//
	// The first is the case where we don't have a batch change yet (we're
	// applying a new batch spec), and some changeset specs have associated
	// publication states. We should get the appropriate actions on those
	// changeset specs.
	//
	// The second is the case where we do have a batch change, and we're
	// updating some publication states. Again, we should get the appropriate
	// actions.
	//
	// Finally, we need to ensure that providing a conflicting UI publication
	// state results in an error.
	//
	// As ever, let's start with some boilerplate.
	if testing.Short() {
		t.Skip()
	}

	ctx := actor.WithInternalActor(context.Background())
	db := dbtest.NewDB(t, "")

	userID := ct.CreateTestUser(t, db, false).ID

	bstore := store.New(db, &observation.TestContext, nil)
	esStore := database.ExternalServicesWith(bstore)
	repoStore := database.ReposWith(bstore)

	repo := newGitHubTestRepo("github.com/sourcegraph/test", newGitHubExternalService(t, esStore))
	if err := repoStore.Create(ctx, repo); err != nil {
		t.Fatal(err)
	}

	s, err := graphqlbackend.NewSchema(db, &Resolver{store: bstore}, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// We need a batch spec and a set of changeset specs that we can use to
	// verify that the behaviour is as expected. We'll create one changeset spec
	// with an explicit published field (so we can verify that UI publication
	// states can't override that), and four changeset specs without published
	// fields (one for each possible publication state).
	var (
		batchSpec = ct.CreateBatchSpec(t, ctx, bstore, "batch-spec", userID)

		specPublished = ct.CreateChangesetSpec(t, ctx, bstore, ct.TestSpecOpts{
			BatchSpec: batchSpec.ID,
			User:      userID,
			Repo:      repo.ID,
			HeadRef:   "published",
			Published: true,
		})
		specToBePublished = ct.CreateChangesetSpec(t, ctx, bstore, ct.TestSpecOpts{
			BatchSpec: batchSpec.ID,
			User:      userID,
			Repo:      repo.ID,
			HeadRef:   "to be published",
		})
		specToBeDraft = ct.CreateChangesetSpec(t, ctx, bstore, ct.TestSpecOpts{
			BatchSpec: batchSpec.ID,
			User:      userID,
			Repo:      repo.ID,
			HeadRef:   "to be draft",
		})
		specToBeUnpublished = ct.CreateChangesetSpec(t, ctx, bstore, ct.TestSpecOpts{
			BatchSpec: batchSpec.ID,
			User:      userID,
			Repo:      repo.ID,
			HeadRef:   "to be unpublished",
		})
		specToBeOmitted = ct.CreateChangesetSpec(t, ctx, bstore, ct.TestSpecOpts{
			BatchSpec: batchSpec.ID,
			User:      userID,
			Repo:      repo.ID,
			HeadRef:   "to be omitted",
		})
	)

	// Most of the input variables are common in the GraphQL queries below, so
	// here's a function to decorate the common input variables with whatever
	// custom variables are required.
	decorateInput := func(in map[string]interface{}) map[string]interface{} {
		commonInputs := map[string]interface{}{
			"batchSpec": marshalBatchSpecRandID(batchSpec.RandID),
			"publicationStates": []map[string]interface{}{
				{
					"changesetSpec":    marshalChangesetSpecRandID(specToBePublished.RandID),
					"publicationState": true,
				},
				{
					"changesetSpec":    marshalChangesetSpecRandID(specToBeDraft.RandID),
					"publicationState": "draft",
				},
				{
					"changesetSpec":    marshalChangesetSpecRandID(specToBeUnpublished.RandID),
					"publicationState": false,
				},
				// We'll also toss in a spec that doesn't exist, since
				// applyPreview() is documented to ignore unknown changeset specs
				// due to its pagination behaviour.
				{
					"changesetSpec":    marshalChangesetSpecRandID("this is not a valid random ID"),
					"publicationState": true,
				},
			},
		}

		for k, v := range in {
			commonInputs[k] = v
		}

		return commonInputs
	}

	// To make it easier to assert against the operations in a preview node,
	// here are some canned operations that we expect when publishing.
	var (
		publishOps = []btypes.ReconcilerOperation{
			btypes.ReconcilerOperationPush,
			btypes.ReconcilerOperationPublish,
		}
		publishDraftOps = []btypes.ReconcilerOperation{
			btypes.ReconcilerOperationPush,
			btypes.ReconcilerOperationPublishDraft,
		}
		noOps = []btypes.ReconcilerOperation{}
	)

	t.Run("new batch change", func(t *testing.T) {
		// We'll use a page size of 1 here to ensure that the publication states
		// are correctly handled across pages.
		previews := repeatApplyPreview(
			ctx, t, s,
			decorateInput(map[string]interface{}{}),
			queryChangesetApplyPreview,
			1,
		)

		assert.Len(t, previews, 5)
		assertOperations(t, previews, specPublished, publishOps)
		assertOperations(t, previews, specToBePublished, publishOps)
		assertOperations(t, previews, specToBeDraft, publishDraftOps)
		assertOperations(t, previews, specToBeUnpublished, noOps)
		assertOperations(t, previews, specToBeOmitted, noOps)
	})

	t.Run("existing batch change", func(t *testing.T) {
		batchChange := ct.CreateBatchChange(t, ctx, bstore, "batch-spec", userID, batchSpec.ID)
		assert.NotNil(t, batchChange)

		// Same as above, but this time we'll use a page size of 2 just to mix
		// it up.
		previews := repeatApplyPreview(
			ctx, t, s,
			decorateInput(map[string]interface{}{}),
			queryChangesetApplyPreview,
			2,
		)

		assert.Len(t, previews, 5)
		assertOperations(t, previews, specPublished, publishOps)
		assertOperations(t, previews, specToBePublished, publishOps)
		assertOperations(t, previews, specToBeDraft, publishDraftOps)
		assertOperations(t, previews, specToBeUnpublished, noOps)
		assertOperations(t, previews, specToBeOmitted, noOps)
	})

	t.Run("conflicting publication state", func(t *testing.T) {
		var response struct{ Node apitest.BatchSpec }
		err := apitest.Exec(
			ctx, t, s,
			decorateInput(map[string]interface{}{
				"publicationStates": []map[string]interface{}{
					{
						"changesetSpec":    marshalChangesetSpecRandID(specPublished.RandID),
						"publicationState": true,
					},
				},
			}),
			&response,
			queryChangesetApplyPreview,
		)

		assert.Len(t, err, 1)
		assert.Error(t, err[0])
	})
}

// assertOperations asserts that the given operations appear for the given
// changeset spec within the array of preview nodes.
func assertOperations(
	t *testing.T,
	previews []apitest.ChangesetApplyPreview,
	spec *btypes.ChangesetSpec,
	want []btypes.ReconcilerOperation,

) {
	t.Helper()

	preview := findPreviewForChangesetSpec(previews, spec)
	if preview == nil {
		t.Fatal("could not find changeset spec")
	}

	assert.Equal(t, want, preview.Operations)
}

func findPreviewForChangesetSpec(
	previews []apitest.ChangesetApplyPreview,
	spec *btypes.ChangesetSpec,
) *apitest.ChangesetApplyPreview {
	id := string(marshalChangesetSpecRandID(spec.RandID))
	for _, preview := range previews {
		if preview.Targets.ChangesetSpec.ID == id {
			return &preview
		}
	}

	return nil
}

// repeatApplyPreview tests the applyPreview resolver's pagination behaviour by
// retrieving the entire set of previews for the given input by making repeated
// requests.
func repeatApplyPreview(
	ctx context.Context,
	t *testing.T,
	schema *graphql.Schema,
	in map[string]interface{},
	query string,
	pageSize int,
) []apitest.ChangesetApplyPreview {
	t.Helper()

	in["first"] = pageSize
	in["after"] = nil
	out := []apitest.ChangesetApplyPreview{}

	for {
		var response struct{ Node apitest.BatchSpec }
		apitest.MustExec(ctx, t, schema, in, &response, query)
		out = append(out, response.Node.ApplyPreview.Nodes...)

		if response.Node.ApplyPreview.PageInfo.HasNextPage {
			in["after"] = *response.Node.ApplyPreview.PageInfo.EndCursor
		} else {
			return out
		}
	}
}
