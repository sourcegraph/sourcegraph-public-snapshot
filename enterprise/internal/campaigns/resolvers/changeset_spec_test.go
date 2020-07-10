package resolvers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestChangesetSpecResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global, "changeset-spec-by-id", false)

	store := ee.NewStore(dbconn.Global)
	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	repo := newGitHubTestRepo("github.com/sourcegraph/sourcegraph", 1)
	if err := reposStore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}
	repoID := graphqlbackend.MarshalRepositoryID(repo.ID)

	s, err := graphqlbackend.NewSchema(&Resolver{store: store}, nil, nil)
	if err != nil {
		t.Fatal(err)

	}

	tests := []struct {
		name    string
		rawSpec string
		want    func(spec *campaigns.ChangesetSpec) apitest.ChangesetSpec
	}{
		{
			name:    "GitBranchChangesetDescription",
			rawSpec: ct.NewRawChangesetSpecGitBranch(repoID),
			want: func(spec *campaigns.ChangesetSpec) apitest.ChangesetSpec {
				return apitest.ChangesetSpec{
					Typename: "ChangesetSpec",
					ID:       string(marshalChangesetSpecRandID(spec.RandID)),
					Description: apitest.ChangesetSpecDescription{
						Typename:       "GitBranchChangesetDescription",
						BaseRepository: string(spec.Spec.BaseRepository),
						ExternalID:     "",
						BaseRef:        spec.Spec.BaseRef,
						HeadRepository: string(spec.Spec.HeadRepository),
						HeadRef:        spec.Spec.HeadRef,
						Title:          spec.Spec.Title,
						Body:           spec.Spec.Body,
						Commits: []apitest.GitCommitDescription{
							{Diff: spec.Spec.Commits[0].Diff, Message: spec.Spec.Commits[0].Message},
						},
						Published: false,
					},
					ExpiresAt: &graphqlbackend.DateTime{
						Time: spec.CreatedAt.Truncate(time.Second).Add(2 * time.Hour),
					},
				}
			},
		},
		{
			name:    "ExistingChangesetReference",
			rawSpec: ct.NewRawChangesetSpecExisting(repoID, "9999"),
			want: func(spec *campaigns.ChangesetSpec) apitest.ChangesetSpec {
				return apitest.ChangesetSpec{
					Typename: "ChangesetSpec",
					ID:       string(marshalChangesetSpecRandID(spec.RandID)),
					Description: apitest.ChangesetSpecDescription{
						Typename:       "ExistingChangesetReference",
						BaseRepository: string(spec.Spec.BaseRepository),
						ExternalID:     spec.Spec.ExternalID,
						Published:      false,
					},
					ExpiresAt: &graphqlbackend.DateTime{
						Time: spec.CreatedAt.Truncate(time.Second).Add(2 * time.Hour),
					},
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := campaigns.NewChangesetSpecFromRaw(tc.rawSpec)
			if err != nil {
				t.Fatal(err)
			}
			spec.UserID = userID
			spec.RepoID = repo.ID

			if err := store.CreateChangesetSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			input := map[string]interface{}{"id": marshalChangesetSpecRandID(spec.RandID)}
			var response struct{ Node apitest.ChangesetSpec }
			apitest.MustExec(ctx, t, s, input, &response, queryChangesetSpecNode)

			want := tc.want(spec)
			if diff := cmp.Diff(want, response.Node); diff != "" {
				t.Fatalf("unexpected response (-want +got):\n%s", diff)
			}
		})
	}
}

const queryChangesetSpecNode = `
query($id: ID!) {
  node(id: $id) {
    __typename

    ... on ChangesetSpec {
      id

      description {
        __typename

        ... on ExistingChangesetReference {
          baseRepository
          externalID
        }

        ... on GitBranchChangesetDescription {
          baseRepository
          baseRef
          baseRev

          headRepository
          headRef

          title
          body

          commits {
            message
            diff
          }

          published
        }
      }

      expiresAt
	}
  }
}
`
