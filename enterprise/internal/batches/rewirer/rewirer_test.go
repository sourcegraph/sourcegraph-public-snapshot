package rewirer

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRewirer_Rewire(t *testing.T) {
	testCampaignID := int64(123)
	testChangesetSpecID := int64(512)
	testRepoID := api.RepoID(128)
	testRepo := &types.Repo{
		ID: testRepoID,
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: extsvc.TypeGitHub,
		},
	}
	unsupportedTestRepoID := api.RepoID(256)
	unsupportedTestRepo := &types.Repo{
		ID: unsupportedTestRepoID,
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: extsvc.TypeBitbucketCloud,
		},
	}
	testCases := []struct {
		name           string
		mappings       store.RewirerMappings
		wantChangesets []ct.ChangesetAssertions
		wantErr        error
	}{
		{
			name:           "empty mappings",
			mappings:       store.RewirerMappings{},
			wantChangesets: []ct.ChangesetAssertions{},
		},
		// NO CHANGESET SPEC
		{
			name: "no spec matching existing imported changeset",
			mappings: store.RewirerMappings{{
				Changeset: ct.BuildChangeset(ct.TestChangesetOpts{
					Repo:      testRepoID,
					Campaigns: []batches.BatchChangeAssoc{{BatchChangeID: testCampaignID}},

					// Imported changeset:
					OwnedByCampaign: 0,
					CurrentSpec:     0,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []ct.ChangesetAssertions{
				// No match, should be re-enqueued and detached from the campaign.
				assertResetQueued(ct.ChangesetAssertions{
					Repo:       testRepoID,
					DetachFrom: []int64{testCampaignID},
				}),
			},
		},
		{
			name: "no spec matching existing unpublished branch changeset owned by this campaign",
			mappings: store.RewirerMappings{{
				Changeset: ct.BuildChangeset(ct.TestChangesetOpts{
					Repo:      testRepoID,
					Campaigns: []batches.BatchChangeAssoc{{BatchChangeID: testCampaignID}},

					// Owned unpublished branch changeset:
					PublicationState: batches.ChangesetPublicationStateUnpublished,
					OwnedByCampaign:  testCampaignID,
					CurrentSpec:      testChangesetSpecID,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []ct.ChangesetAssertions{
				// No match, should be re-enqueued and detached from the campaign.
				assertResetQueued(ct.ChangesetAssertions{
					PublicationState: batches.ChangesetPublicationStateUnpublished,
					OwnedByCampaign:  testCampaignID,
					CurrentSpec:      testChangesetSpecID,
					Repo:             testRepoID,
					DetachFrom:       []int64{testCampaignID},
				}),
			},
		},
		{
			name: "no spec matching existing published branch changeset owned by this campaign",
			mappings: store.RewirerMappings{{
				Changeset: ct.BuildChangeset(ct.TestChangesetOpts{
					Repo:      testRepoID,
					Campaigns: []batches.BatchChangeAssoc{{BatchChangeID: testCampaignID}},

					// Owned, published branch changeset:
					OwnedByCampaign:  testCampaignID,
					CurrentSpec:      testChangesetSpecID,
					PublicationState: batches.ChangesetPublicationStatePublished,
					// Publication succeeded
					ReconcilerState: batches.ReconcilerStateCompleted,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []ct.ChangesetAssertions{
				// No match, should be re-enqueued and detached from the campaign.
				assertResetQueued(ct.ChangesetAssertions{
					PublicationState: batches.ChangesetPublicationStatePublished,
					OwnedByCampaign:  testCampaignID,
					CurrentSpec:      testChangesetSpecID,
					// The changeset should be closed on the code host.
					Closing:    true,
					Repo:       testRepoID,
					DetachFrom: []int64{testCampaignID},
					// Current spec should have been made the previous spec.
					PreviousSpec: testChangesetSpecID,
				}),
			},
		},
		{
			name: "no spec matching existing changeset, no repo perms",
			mappings: store.RewirerMappings{{
				Changeset: ct.BuildChangeset(ct.TestChangesetOpts{
					Repo:      0,
					Campaigns: []batches.BatchChangeAssoc{{BatchChangeID: testCampaignID}},
				}),
				// No access to repo.
				Repo: nil,
			}},
			// Nothing should be done.
			wantChangesets: []ct.ChangesetAssertions{},
		},
		// END NO CHANGESET SPEC
		// NO CHANGESET
		{
			name: "new importing spec",
			mappings: store.RewirerMappings{{
				ChangesetSpec: ct.BuildChangesetSpec(t, ct.TestSpecOpts{
					Repo: testRepoID,

					// Importing spec
					ExternalID: "123",
				}),
				Repo: testRepo,
			}},
			wantChangesets: []ct.ChangesetAssertions{assertResetQueued(ct.ChangesetAssertions{
				Repo:       testRepoID,
				ExternalID: "123",
				// Imported changesets always start as unpublished and will be set to published once the import succeeded.
				PublicationState: batches.ChangesetPublicationStateUnpublished,
				AttachedTo:       []int64{testCampaignID},
			})},
		},
		{
			name: "new branch spec",
			mappings: store.RewirerMappings{{
				ChangesetSpec: ct.BuildChangesetSpec(t, ct.TestSpecOpts{
					ID:   testChangesetSpecID,
					Repo: testRepoID,

					// Branch spec
					HeadRef: "refs/heads/test-branch",
				}),
				Repo: testRepo,
			}},
			wantChangesets: []ct.ChangesetAssertions{assertResetQueued(ct.ChangesetAssertions{
				Repo:             testRepoID,
				PublicationState: batches.ChangesetPublicationStateUnpublished,
				AttachedTo:       []int64{testCampaignID},
				OwnedByCampaign:  testCampaignID,
				CurrentSpec:      testChangesetSpecID,
				// Diff stat is copied over from changeset spec
				DiffStat: ct.TestChangsetSpecDiffStat,
			})},
		},
		{
			name: "unsupported repo",
			mappings: store.RewirerMappings{{
				ChangesetSpec: ct.BuildChangesetSpec(t, ct.TestSpecOpts{
					Repo:       unsupportedTestRepoID,
					ExternalID: "123",
				}),
				RepoID: unsupportedTestRepoID,
				Repo:   unsupportedTestRepo,
			}},
			wantErr: &ErrRepoNotSupported{
				ServiceType: unsupportedTestRepo.ExternalRepo.ServiceType,
				RepoName:    string(unsupportedTestRepo.Name),
			},
		},
		{
			name: "inaccessible repo",
			mappings: store.RewirerMappings{{
				ChangesetSpec: ct.BuildChangesetSpec(t, ct.TestSpecOpts{
					Repo:       testRepoID,
					ExternalID: "123",
				}),
				RepoID: testRepoID,
				Repo:   nil,
			}},
			wantErr: &database.RepoNotFoundErr{ID: testRepoID},
		},
		// END NO CHANGESET
		// CHANGESET SPEC AND CHANGESET
		{
			name: "update importing spec: imported by other",
			mappings: store.RewirerMappings{{
				ChangesetSpec: ct.BuildChangesetSpec(t, ct.TestSpecOpts{
					Repo: testRepoID,

					// Importing spec
					ExternalID: "123",
				}),
				Changeset: ct.BuildChangeset(ct.TestChangesetOpts{
					Repo:       testRepoID,
					ExternalID: "123",
					// Already attached to another campaign
					Campaigns: []batches.BatchChangeAssoc{{BatchChangeID: testCampaignID + 1}},
				}),
				Repo: testRepo,
			}},
			wantChangesets: []ct.ChangesetAssertions{
				// Should not be reenqueued
				{
					Repo:       testRepoID,
					ExternalID: "123",
					// Now should be attached to both batches.
					AttachedTo: []int64{testCampaignID + 1, testCampaignID},
				},
			},
		},
		{
			name: "update importing spec: failed before",
			mappings: store.RewirerMappings{{
				ChangesetSpec: ct.BuildChangesetSpec(t, ct.TestSpecOpts{
					Repo: testRepoID,

					// Importing spec
					ExternalID: "123",
				}),
				Changeset: ct.BuildChangeset(ct.TestChangesetOpts{
					Repo:       testRepoID,
					ExternalID: "123",
					// Already attached to another campaign
					Campaigns:       []batches.BatchChangeAssoc{{BatchChangeID: testCampaignID + 1}},
					ReconcilerState: batches.ReconcilerStateFailed,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []ct.ChangesetAssertions{assertResetQueued(ct.ChangesetAssertions{
				Repo:       testRepoID,
				ExternalID: "123",
				// Now should be attached to both batches.
				AttachedTo: []int64{testCampaignID + 1, testCampaignID},
			})},
		},
		{
			name: "update importing spec: created by other campaign",
			mappings: store.RewirerMappings{{
				ChangesetSpec: ct.BuildChangesetSpec(t, ct.TestSpecOpts{
					Repo: testRepoID,

					// Importing spec
					ExternalID: "123",
				}),
				Changeset: ct.BuildChangeset(ct.TestChangesetOpts{
					Repo:       testRepoID,
					ExternalID: "123",
					// Already attached to another campaign
					Campaigns: []batches.BatchChangeAssoc{{BatchChangeID: testCampaignID + 1}},
					// Other campaign created this changeset.
					OwnedByCampaign: testCampaignID + 1,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []ct.ChangesetAssertions{
				// Changeset owned by another campaign should not be retried.
				{
					Repo:            testRepoID,
					ExternalID:      "123",
					OwnedByCampaign: testCampaignID + 1,
					// Now should be attached to both batches.
					AttachedTo: []int64{testCampaignID + 1, testCampaignID},
				}},
		},
		{
			name: "update branch spec",
			mappings: store.RewirerMappings{{
				ChangesetSpec: ct.BuildChangesetSpec(t, ct.TestSpecOpts{
					ID:   testChangesetSpecID + 1,
					Repo: testRepoID,

					// Branch spec
					HeadRef: "refs/heads/test-branch",
				}),
				Changeset: ct.BuildChangeset(ct.TestChangesetOpts{
					Repo:             testRepoID,
					ExternalID:       "123",
					CurrentSpec:      testChangesetSpecID,
					Campaigns:        []batches.BatchChangeAssoc{{BatchChangeID: testCampaignID}},
					OwnedByCampaign:  testCampaignID,
					PublicationState: batches.ChangesetPublicationStatePublished,
					ReconcilerState:  batches.ReconcilerStateCompleted,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []ct.ChangesetAssertions{assertResetQueued(ct.ChangesetAssertions{
				Repo:             testRepoID,
				ExternalID:       "123",
				OwnedByCampaign:  testCampaignID,
				AttachedTo:       []int64{testCampaignID},
				PublicationState: batches.ChangesetPublicationStatePublished,
				CurrentSpec:      testChangesetSpecID + 1,
				// The changeset was reconciled successfully before, so the previous spec should have been recorded.
				PreviousSpec: testChangesetSpecID,
			})},
		},
		{
			name: "update branch spec - failed before",
			mappings: store.RewirerMappings{{
				ChangesetSpec: ct.BuildChangesetSpec(t, ct.TestSpecOpts{
					ID:   testChangesetSpecID + 1,
					Repo: testRepoID,

					// Branch spec
					HeadRef: "refs/heads/test-branch",
				}),
				Changeset: ct.BuildChangeset(ct.TestChangesetOpts{
					Repo:             testRepoID,
					ExternalID:       "123",
					CurrentSpec:      testChangesetSpecID,
					Campaigns:        []batches.BatchChangeAssoc{{BatchChangeID: testCampaignID}},
					OwnedByCampaign:  testCampaignID,
					PublicationState: batches.ChangesetPublicationStatePublished,
					ReconcilerState:  batches.ReconcilerStateFailed,
				}),
				Repo: testRepo,
			}},
			wantChangesets: []ct.ChangesetAssertions{assertResetQueued(ct.ChangesetAssertions{
				Repo:             testRepoID,
				ExternalID:       "123",
				OwnedByCampaign:  testCampaignID,
				AttachedTo:       []int64{testCampaignID},
				PublicationState: batches.ChangesetPublicationStatePublished,
				CurrentSpec:      testChangesetSpecID + 1,
				// The changeset was not reconciled successfully before, so the previous spec should have remained unset.
				PreviousSpec: 0,
			})},
		},
		// END CHANGESET SPEC AND CHANGESET
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := New(tc.mappings, testCampaignID)
			changesets, err := r.Rewire()
			if err != nil && tc.wantErr == nil {
				t.Fatal(err)
			}
			if tc.wantErr != nil && err.Error() != tc.wantErr.Error() {
				t.Fatalf("incorrect error returned. want=%+v have=%+v", tc.wantErr, err)
			}
			if have, want := len(changesets), len(tc.wantChangesets); have != want {
				t.Fatalf("incorrect amount of changesets returned. want=%d have=%d", want, have)
			}
			for i, changeset := range changesets {
				ct.AssertChangeset(t, changeset, tc.wantChangesets[i])
			}
		})
	}
}

func assertResetQueued(a ct.ChangesetAssertions) ct.ChangesetAssertions {
	a.ReconcilerState = batches.ReconcilerStateQueued
	a.NumFailures = 0
	a.NumResets = 0
	a.FailureMessage = nil
	a.SyncErrorMessage = nil
	return a
}
