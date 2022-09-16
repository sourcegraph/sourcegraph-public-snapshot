package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/batches"
)

func TestChangesetSpecResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CreateTestUser(t, db, false).ID

	bstore := store.New(db, &observation.TestContext, nil)
	esStore := database.ExternalServicesWith(logger, bstore)

	// Creating user with matching email to the changeset spec author.
	user, err := database.UsersWith(logger, bstore).Create(ctx, database.NewUser{
		Username:        "mary",
		Email:           bt.ChangesetSpecAuthorEmail,
		EmailIsVerified: true,
		DisplayName:     "Mary Tester",
	})
	if err != nil {
		t.Fatal(err)
	}

	repoStore := database.ReposWith(logger, bstore)
	repo := newGitHubTestRepo("github.com/sourcegraph/changeset-spec-resolver-test", newGitHubExternalService(t, esStore))
	if err := repoStore.Create(ctx, repo); err != nil {
		t.Fatal(err)
	}
	repoID := graphqlbackend.MarshalRepositoryID(repo.ID)

	testRev := api.CommitID("b69072d5f687b31b9f6ae3ceafdc24c259c4b9ec")
	mockBackendCommits(t, testRev)

	batchSpec, err := btypes.NewBatchSpecFromRaw(`name: awesome-test`)
	if err != nil {
		t.Fatal(err)
	}
	batchSpec.NamespaceUserID = userID
	if err := bstore.CreateBatchSpec(ctx, batchSpec); err != nil {
		t.Fatal(err)
	}

	s, err := newSchema(db, &Resolver{store: bstore})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		rawSpec string
		want    func(spec *btypes.ChangesetSpec) apitest.ChangesetSpec
	}{
		{
			name:    "GitBranchChangesetDescription",
			rawSpec: bt.NewRawChangesetSpecGitBranch(repoID, string(testRev)),
			want: func(spec *btypes.ChangesetSpec) apitest.ChangesetSpec {
				return apitest.ChangesetSpec{
					Typename: "VisibleChangesetSpec",
					ID:       string(marshalChangesetSpecRandID(spec.RandID)),
					Description: apitest.ChangesetSpecDescription{
						Typename: "GitBranchChangesetDescription",
						BaseRepository: apitest.Repository{
							ID: string(graphqlbackend.MarshalRepositoryID(spec.BaseRepoID)),
						},
						ExternalID: "",
						BaseRef:    gitdomain.AbbreviateRef(spec.BaseRef),
						HeadRef:    gitdomain.AbbreviateRef(spec.HeadRef),
						Title:      spec.Title,
						Body:       spec.Body,
						Commits: []apitest.GitCommitDescription{
							{
								Author: apitest.Person{
									Email: spec.CommitAuthorEmail,
									Name:  user.Username,
									User: &apitest.User{
										ID: string(graphqlbackend.MarshalUserID(user.ID)),
									},
								},
								Diff:    string(spec.Diff),
								Message: spec.CommitMessage,
								Subject: "git commit message",
								Body:    "and some more content in a second paragraph.",
							},
						},
						Published: batches.PublishedValue{Val: false},
						Diff: struct{ FileDiffs apitest.FileDiffs }{
							FileDiffs: apitest.FileDiffs{
								DiffStat: apitest.DiffStat{
									Added:   1,
									Deleted: 1,
									Changed: 2,
								},
							},
						},
						DiffStat: apitest.DiffStat{
							Added:   1,
							Deleted: 1,
							Changed: 2,
						},
					},
					ExpiresAt: &graphqlbackend.DateTime{Time: spec.ExpiresAt().Truncate(time.Second)},
				}
			},
		},
		{
			name:    "GitBranchChangesetDescription Draft",
			rawSpec: bt.NewPublishedRawChangesetSpecGitBranch(repoID, string(testRev), batches.PublishedValue{Val: "draft"}),
			want: func(spec *btypes.ChangesetSpec) apitest.ChangesetSpec {
				return apitest.ChangesetSpec{
					Typename: "VisibleChangesetSpec",
					ID:       string(marshalChangesetSpecRandID(spec.RandID)),
					Description: apitest.ChangesetSpecDescription{
						Typename: "GitBranchChangesetDescription",
						BaseRepository: apitest.Repository{
							ID: string(graphqlbackend.MarshalRepositoryID(spec.BaseRepoID)),
						},
						ExternalID: "",
						BaseRef:    gitdomain.AbbreviateRef(spec.BaseRef),
						HeadRef:    gitdomain.AbbreviateRef(spec.HeadRef),
						Title:      spec.Title,
						Body:       spec.Body,
						Commits: []apitest.GitCommitDescription{
							{
								Author: apitest.Person{
									Email: spec.CommitAuthorEmail,
									Name:  user.Username,
									User: &apitest.User{
										ID: string(graphqlbackend.MarshalUserID(user.ID)),
									},
								},
								Diff:    string(spec.Diff),
								Message: spec.CommitMessage,
								Subject: "git commit message",
								Body:    "and some more content in a second paragraph.",
							},
						},
						Published: batches.PublishedValue{Val: "draft"},
						Diff: struct{ FileDiffs apitest.FileDiffs }{
							FileDiffs: apitest.FileDiffs{
								DiffStat: apitest.DiffStat{
									Added:   1,
									Deleted: 1,
									Changed: 2,
								},
							},
						},
						DiffStat: apitest.DiffStat{
							Added:   1,
							Deleted: 1,
							Changed: 2,
						},
					},
					ExpiresAt: &graphqlbackend.DateTime{Time: spec.ExpiresAt().Truncate(time.Second)},
				}
			},
		},
		{
			name:    "GitBranchChangesetDescription publish from UI",
			rawSpec: bt.NewPublishedRawChangesetSpecGitBranch(repoID, string(testRev), batches.PublishedValue{Val: nil}),
			want: func(spec *btypes.ChangesetSpec) apitest.ChangesetSpec {
				return apitest.ChangesetSpec{
					Typename: "VisibleChangesetSpec",
					ID:       string(marshalChangesetSpecRandID(spec.RandID)),
					Description: apitest.ChangesetSpecDescription{
						Typename: "GitBranchChangesetDescription",
						BaseRepository: apitest.Repository{
							ID: string(graphqlbackend.MarshalRepositoryID(spec.BaseRepoID)),
						},
						ExternalID: "",
						BaseRef:    gitdomain.AbbreviateRef(spec.BaseRef),
						HeadRef:    gitdomain.AbbreviateRef(spec.HeadRef),
						Title:      spec.Title,
						Body:       spec.Body,
						Commits: []apitest.GitCommitDescription{
							{
								Author: apitest.Person{
									Email: spec.CommitAuthorEmail,
									Name:  user.Username,
									User: &apitest.User{
										ID: string(graphqlbackend.MarshalUserID(user.ID)),
									},
								},
								Diff:    string(spec.Diff),
								Message: spec.CommitMessage,
								Subject: "git commit message",
								Body:    "and some more content in a second paragraph.",
							},
						},
						Published: batches.PublishedValue{Val: nil},
						Diff: struct{ FileDiffs apitest.FileDiffs }{
							FileDiffs: apitest.FileDiffs{
								DiffStat: apitest.DiffStat{
									Added:   1,
									Deleted: 1,
									Changed: 2,
								},
							},
						},
						DiffStat: apitest.DiffStat{
							Added:   1,
							Deleted: 1,
							Changed: 2,
						},
					},
					ExpiresAt: &graphqlbackend.DateTime{Time: spec.ExpiresAt().Truncate(time.Second)},
				}
			},
		},
		{
			name:    "ExistingChangesetReference",
			rawSpec: bt.NewRawChangesetSpecExisting(repoID, "9999"),
			want: func(spec *btypes.ChangesetSpec) apitest.ChangesetSpec {
				return apitest.ChangesetSpec{
					Typename: "VisibleChangesetSpec",
					ID:       string(marshalChangesetSpecRandID(spec.RandID)),
					Description: apitest.ChangesetSpecDescription{
						Typename: "ExistingChangesetReference",
						BaseRepository: apitest.Repository{
							ID: string(graphqlbackend.MarshalRepositoryID(spec.BaseRepoID)),
						},
						ExternalID: spec.ExternalID,
					},
					ExpiresAt: &graphqlbackend.DateTime{Time: spec.ExpiresAt().Truncate(time.Second)},
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := btypes.NewChangesetSpecFromRaw(tc.rawSpec)
			if err != nil {
				t.Fatal(err)
			}
			spec.UserID = userID
			spec.BaseRepoID = repo.ID
			spec.BatchSpecID = batchSpec.ID

			if err := bstore.CreateChangesetSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			input := map[string]any{"id": marshalChangesetSpecRandID(spec.RandID)}
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

    ... on VisibleChangesetSpec {
      id
      description {
        __typename

        ... on ExistingChangesetReference {
          baseRepository {
             id
          }
          externalID
        }

        ... on GitBranchChangesetDescription {
          baseRepository {
              id
          }
          baseRef
          baseRev

          headRef

          title
          body

          commits {
            message
            subject
            body
            diff
            author {
              name
              email
              user {
                id
              }
            }
          }

          published

          diff {
            fileDiffs {
              diffStat { added, changed, deleted }
            }
          }
          diffStat { added, changed, deleted }
        }
      }

      expiresAt
    }
  }
}
`
