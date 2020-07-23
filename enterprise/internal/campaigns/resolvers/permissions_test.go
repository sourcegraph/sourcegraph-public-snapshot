package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestPermissionLevels(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	dbtesting.SetupGlobalTestDB(t)

	store := ee.NewStore(dbconn.Global)
	sr := &Resolver{store: store}
	s, err := graphqlbackend.NewSchema(sr, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Global test data that we reuse in every test
	adminID := insertTestUser(t, dbconn.Global, "perm-level-admin", true)
	userID := insertTestUser(t, dbconn.Global, "perm-level-user", false)

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	repo := newGitHubTestRepo("github.com/sourcegraph/sourcegraph", 1)
	if err := reposStore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	changeset := &campaigns.Changeset{
		RepoID:              repo.ID,
		ExternalServiceType: "github",
		ExternalID:          "1234",
	}
	if err := store.CreateChangesets(ctx, changeset); err != nil {
		t.Fatal(err)
	}

	createCampaign := func(t *testing.T, s *ee.Store, name string, userID int32) (campaignID int64) {
		t.Helper()

		c := &campaigns.Campaign{
			Name:            name,
			AuthorID:        userID,
			NamespaceUserID: userID,
			// We attach the changeset to the campaign so we can test syncChangeset
			ChangesetIDs: []int64{changeset.ID},
		}
		if err := s.CreateCampaign(ctx, c); err != nil {
			t.Fatal(err)
		}

		cs := &campaigns.CampaignSpec{UserID: userID, NamespaceUserID: userID}
		if err := s.CreateCampaignSpec(ctx, cs); err != nil {
			t.Fatal(err)
		}

		return c.ID
	}

	createCampaignSpec := func(t *testing.T, s *ee.Store, userID int32) (randID string) {
		t.Helper()

		cs := &campaigns.CampaignSpec{UserID: userID, NamespaceUserID: userID}
		if err := s.CreateCampaignSpec(ctx, cs); err != nil {
			t.Fatal(err)
		}

		return cs.RandID
	}

	cleanUpCampaigns := func(t *testing.T, s *ee.Store) {
		t.Helper()

		campaigns, next, err := store.ListCampaigns(ctx, ee.ListCampaignsOpts{Limit: 1000})
		if err != nil {
			t.Fatal(err)
		}
		if next != 0 {
			t.Fatalf("more campaigns in store")
		}

		for _, c := range campaigns {
			if err := store.DeleteCampaign(ctx, c.ID); err != nil {
				t.Fatal(err)
			}
		}
	}

	t.Run("queries", func(t *testing.T) {
		// We need to enable read access so that non-site-admin users can access
		// the API and we can check for their admin rights.
		// This can be removed once we enable campaigns for all users and only
		// check for permissions.
		readAccessEnabled := true
		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
			CampaignsReadAccessEnabled: &readAccessEnabled,
		}})
		defer conf.Mock(nil)

		cleanUpCampaigns(t, store)

		adminCampaign := createCampaign(t, store, "admin", adminID)
		adminCampaignSpec := createCampaignSpec(t, store, adminID)
		userCampaign := createCampaign(t, store, "user", userID)
		userCampaignSpec := createCampaignSpec(t, store, userID)

		t.Run("CampaignByID", func(t *testing.T) {
			tests := []struct {
				name                    string
				currentUser             int32
				campaign                int64
				wantViewerCanAdminister bool
			}{
				{
					name:                    "site-admin viewing own campaign",
					currentUser:             adminID,
					campaign:                adminCampaign,
					wantViewerCanAdminister: true,
				},
				{
					name:                    "non-site-admin viewing other's campaign",
					currentUser:             userID,
					campaign:                adminCampaign,
					wantViewerCanAdminister: false,
				},
				{
					name:                    "site-admin viewing other's campaign",
					currentUser:             adminID,
					campaign:                userCampaign,
					wantViewerCanAdminister: true,
				},
				{
					name:                    "non-site-admin viewing own campaign",
					currentUser:             userID,
					campaign:                userCampaign,
					wantViewerCanAdminister: true,
				},
			}

			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					graphqlID := string(campaigns.MarshalCampaignID(tc.campaign))

					var res struct{ Node apitest.Campaign }

					input := map[string]interface{}{"campaign": graphqlID}
					queryCampaign := `
				  query($campaign: ID!) {
				    node(id: $campaign) { ... on Campaign { id, viewerCanAdminister } }
				  }
                `

					actorCtx := actor.WithActor(ctx, actor.FromUser(tc.currentUser))
					apitest.MustExec(actorCtx, t, s, input, &res, queryCampaign)

					if have, want := res.Node.ID, graphqlID; have != want {
						t.Fatalf("queried campaign has wrong id %q, want %q", have, want)
					}
					if have, want := res.Node.ViewerCanAdminister, tc.wantViewerCanAdminister; have != want {
						t.Fatalf("queried campaign's ViewerCanAdminister is wrong %t, want %t", have, want)
					}
				})
			}
		})

		t.Run("CampaignSpecByID", func(t *testing.T) {
			tests := []struct {
				name                    string
				currentUser             int32
				campaignSpec            string
				wantViewerCanAdminister bool
			}{
				{
					name:                    "site-admin viewing own campaign spec",
					currentUser:             adminID,
					campaignSpec:            adminCampaignSpec,
					wantViewerCanAdminister: true,
				},
				{
					name:                    "non-site-admin viewing other's campaign spec",
					currentUser:             userID,
					campaignSpec:            adminCampaignSpec,
					wantViewerCanAdminister: false,
				},
				{
					name:                    "site-admin viewing other's campaign spec",
					currentUser:             adminID,
					campaignSpec:            userCampaignSpec,
					wantViewerCanAdminister: true,
				},
				{
					name:                    "non-site-admin viewing own campaign spec",
					currentUser:             userID,
					campaignSpec:            userCampaignSpec,
					wantViewerCanAdminister: true,
				},
			}

			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					graphqlID := string(marshalCampaignSpecRandID(tc.campaignSpec))

					var res struct{ Node apitest.CampaignSpec }

					input := map[string]interface{}{"campaignSpec": graphqlID}
					queryCampaignSpec := `
				  query($campaignSpec: ID!) {
				    node(id: $campaignSpec) { ... on CampaignSpec { id, viewerCanAdminister } }
				  }
                `

					actorCtx := actor.WithActor(ctx, actor.FromUser(tc.currentUser))
					apitest.MustExec(actorCtx, t, s, input, &res, queryCampaignSpec)

					if have, want := res.Node.ID, graphqlID; have != want {
						t.Fatalf("queried campaign spec has wrong id %q, want %q", have, want)
					}
					if have, want := res.Node.ViewerCanAdminister, tc.wantViewerCanAdminister; have != want {
						t.Fatalf("queried campaign spec's ViewerCanAdminister is wrong %t, want %t", have, want)
					}
				})
			}
		})

		t.Run("Campaigns", func(t *testing.T) {
			tests := []struct {
				name                string
				currentUser         int32
				viewerCanAdminister bool
				wantCampaigns       []int64
			}{
				{
					name:                "admin listing viewerCanAdminister: true",
					currentUser:         adminID,
					viewerCanAdminister: true,
					wantCampaigns:       []int64{adminCampaign, userCampaign},
				},
				{
					name:                "user listing viewerCanAdminister: true",
					currentUser:         userID,
					viewerCanAdminister: true,
					wantCampaigns:       []int64{userCampaign},
				},
				{
					name:                "admin listing viewerCanAdminister: false",
					currentUser:         adminID,
					viewerCanAdminister: false,
					wantCampaigns:       []int64{adminCampaign, userCampaign},
				},
				{
					name:                "user listing viewerCanAdminister: false",
					currentUser:         userID,
					viewerCanAdminister: false,
					wantCampaigns:       []int64{adminCampaign, userCampaign},
				},
			}
			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					actorCtx := actor.WithActor(context.Background(), actor.FromUser(tc.currentUser))
					expectedIDs := make(map[string]bool, len(tc.wantCampaigns))
					for _, c := range tc.wantCampaigns {
						graphqlID := string(campaigns.MarshalCampaignID(c))
						expectedIDs[graphqlID] = true
					}

					query := fmt.Sprintf(`
				query {
					campaigns(viewerCanAdminister: %t) { totalCount, nodes { id } }
					node(id: %q) {
						id
						... on ExternalChangeset {
							campaigns(viewerCanAdminister: %t) { totalCount, nodes { id } }
						}
					}
					}`, tc.viewerCanAdminister, marshalChangesetID(changeset.ID), tc.viewerCanAdminister)
					var res struct {
						Campaigns apitest.CampaignConnection
						Node      apitest.Changeset
					}
					apitest.MustExec(actorCtx, t, s, nil, &res, query)
					for _, conn := range []apitest.CampaignConnection{res.Campaigns, res.Node.Campaigns} {
						if have, want := conn.TotalCount, len(tc.wantCampaigns); have != want {
							t.Fatalf("wrong count of campaigns returned, want=%d have=%d", want, have)
						}
						if have, want := conn.TotalCount, len(conn.Nodes); have != want {
							t.Fatalf("totalCount and nodes length don't match, want=%d have=%d", want, have)
						}
						for _, node := range conn.Nodes {
							if _, ok := expectedIDs[node.ID]; !ok {
								t.Fatalf("received wrong campaign with id %q", node.ID)
							}
						}
					}
				})
			}
		})
	})

	t.Run("mutations", func(t *testing.T) {
		mutations := []struct {
			name         string
			mutationFunc func(campaignID, changesetID, campaignSpecID string) string
		}{
			{
				name: "closeCampaign",
				mutationFunc: func(campaignID, changesetID, campaignSpecID string) string {
					return fmt.Sprintf(`mutation { closeCampaign(campaign: %q, closeChangesets: false) { id } }`, campaignID)
				},
			},
			{
				name: "deleteCampaign",
				mutationFunc: func(campaignID, changesetID, campaignSpecID string) string {
					return fmt.Sprintf(`mutation { deleteCampaign(campaign: %q) { alwaysNil } } `, campaignID)
				},
			},
			{
				name: "syncChangeset",
				mutationFunc: func(campaignID, changesetID, campaignSpecID string) string {
					return fmt.Sprintf(`mutation { syncChangeset(changeset: %q) { alwaysNil } }`, changesetID)
				},
			},
			{
				name: "applyCampaign",
				mutationFunc: func(campaignID, changesetID, campaignSpecID string) string {
					return fmt.Sprintf(`mutation { applyCampaign(campaignSpec: %q) { id } }`, campaignSpecID)
				},
			},
			{
				name: "moveCampaign",
				mutationFunc: func(campaignID, changesetID, campaignSpecID string) string {
					return fmt.Sprintf(`mutation { moveCampaign(campaign: %q, newName: "foobar") { id } }`, campaignID)
				},
			},
		}

		for _, m := range mutations {
			t.Run(m.name, func(t *testing.T) {
				tests := []struct {
					name           string
					currentUser    int32
					campaignAuthor int32
					wantAuthErr    bool
				}{
					{
						name:           "unauthorized",
						currentUser:    userID,
						campaignAuthor: adminID,
						wantAuthErr:    true,
					},
					{
						name:           "authorized campaign owner",
						currentUser:    userID,
						campaignAuthor: userID,
						wantAuthErr:    false,
					},
					{
						name:           "authorized site-admin",
						currentUser:    adminID,
						campaignAuthor: userID,
						wantAuthErr:    false,
					},
				}

				for _, tc := range tests {
					t.Run(tc.name, func(t *testing.T) {
						cleanUpCampaigns(t, store)

						campaignID := createCampaign(t, store, "test-campaign", tc.campaignAuthor)
						campaignSpecID := createCampaignSpec(t, store, tc.campaignAuthor)

						// We add the changeset to the campaign. It doesn't matter
						// for the addChangesetsToCampaign mutation, since that is
						// idempotent and we want to solely check for auth errors.
						changeset.CampaignIDs = []int64{campaignID}
						if err := store.UpdateChangesets(ctx, changeset); err != nil {
							t.Fatal(err)
						}

						mutation := m.mutationFunc(
							string(campaigns.MarshalCampaignID(campaignID)),
							string(marshalChangesetID(changeset.ID)),
							string(marshalCampaignSpecRandID(campaignSpecID)),
						)

						actorCtx := actor.WithActor(ctx, actor.FromUser(tc.currentUser))

						var response struct{}
						errs := apitest.Exec(actorCtx, t, s, nil, &response, mutation)

						if tc.wantAuthErr {
							if len(errs) != 1 {
								t.Fatalf("expected 1 error, but got %d: %s", len(errs), errs)
							}
							if !strings.Contains(errs[0].Error(), "must be authenticated") {
								t.Fatalf("wrong error: %s %T", errs[0], errs[0])
							}
						} else {
							// We don't care about other errors, we only want to
							// check that we didn't get an auth error.
							for _, e := range errs {
								if strings.Contains(e.Error(), "must be authenticated") {
									t.Fatalf("auth error wrongly returned: %s %T", errs[0], errs[0])
								}
							}
						}
					})
				}
			})
		}
	})
}

func TestRepositoryPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// now := time.Now().UTC().Truncate(time.Microsecond)

	// We need to enable read access so that non-site-admin users can access
	// the API and we can check for their admin rights.
	// This can be removed once we enable campaigns for all users and only
	// check for permissions.
	readAccessEnabled := true
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		CampaignsReadAccessEnabled: &readAccessEnabled,
	}})
	defer conf.Mock(nil)

	dbtesting.SetupGlobalTestDB(t)

	store := ee.NewStore(dbconn.Global)
	sr := &Resolver{store: store}
	s, err := graphqlbackend.NewSchema(sr, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	testRev := api.CommitID("b69072d5f687b31b9f6ae3ceafdc24c259c4b9ec")
	mockBackendCommits(t, testRev)

	// Global test data that we reuse in every test
	userID := insertTestUser(t, dbconn.Global, "perm-level-user", false)

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	// Create 2 repositories
	repos := make([]*repos.Repo, 0, 2)
	for i := 0; i < cap(repos); i++ {
		name := fmt.Sprintf("github.com/sourcegraph/repo-%d", i)
		r := newGitHubTestRepo(name, i)
		if err := reposStore.UpsertRepos(ctx, r); err != nil {
			t.Fatal(err)
		}
		repos = append(repos, r)
	}

	t.Run("Campaign and changesets", func(t *testing.T) {
		// Create 2 changesets for 2 repositories
		changesetBaseRefOid := "f00b4r"
		changesetHeadRefOid := "b4rf00"
		mockRepoComparison(t, changesetBaseRefOid, changesetHeadRefOid, testDiff)
		changesetDiffStat := apitest.DiffStat{Added: 0, Changed: 2, Deleted: 0}

		changesets := make([]*campaigns.Changeset, 0, len(repos))
		changesetIDs := make([]int64, 0, cap(changesets))
		for _, r := range repos {
			c := &campaigns.Changeset{
				RepoID:              r.ID,
				ExternalServiceType: extsvc.TypeGitHub,
				ExternalID:          fmt.Sprintf("external-%d", r.ID),
				ExternalState:       campaigns.ChangesetExternalStateOpen,
				ExternalCheckState:  campaigns.ChangesetCheckStatePassed,
				ExternalReviewState: campaigns.ChangesetReviewStateChangesRequested,
				PublicationState:               campaigns.ChangesetPublicationStatePublished,
				ReconcilerState:         campaigns.ReconcilerStateCompleted,
				Metadata: &github.PullRequest{
					BaseRefOid: changesetBaseRefOid,
					HeadRefOid: changesetHeadRefOid,
				},
			}
			c.SetDiffStat(changesetDiffStat.ToDiffStat())
			if err := store.CreateChangesets(ctx, c); err != nil {
				t.Fatal(err)
			}
			changesets = append(changesets, c)
			changesetIDs = append(changesetIDs, c.ID)
		}

		campaign := &campaigns.Campaign{
			Name:            "my campaign",
			AuthorID:        userID,
			NamespaceUserID: userID,
			// We attach the two changesets to the campaign
			ChangesetIDs: changesetIDs,
		}
		if err := store.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}
		for _, c := range changesets {
			c.CampaignIDs = []int64{campaign.ID}
		}
		if err := store.UpdateChangesets(ctx, changesets...); err != nil {
			t.Fatal(err)
		}

		// Query campaign and check that we get all changesets
		userCtx := actor.WithActor(ctx, actor.FromUser(userID))

		input := map[string]interface{}{
			"campaign": string(campaigns.MarshalCampaignID(campaign.ID)),
		}
		testCampaignResponse(t, s, userCtx, input, wantCampaignResponse{
			changesetTypes:  map[string]int{"ExternalChangeset": 2},
			changesetsCount: 2,
			campaignDiffStat: apitest.DiffStat{
				Added:   2 * changesetDiffStat.Added,
				Changed: 2 * changesetDiffStat.Changed,
				Deleted: 2 * changesetDiffStat.Deleted,
			},
		})

		for _, c := range changesets {
			// Both changesets are visible still, so both should be ExternalChangesets
			testChangesetResponse(t, s, userCtx, c.ID, "ExternalChangeset")
		}

		// Now we add the authzFilter and filter out the repository of one changeset
		filteredRepo := changesets[0].RepoID
		ct.AuthzFilterRepos(t, filteredRepo)

		// Send query again and check that for each filtered repository we get a
		// HiddenChangeset
		want := wantCampaignResponse{
			changesetTypes: map[string]int{
				"ExternalChangeset":       1,
				"HiddenExternalChangeset": 1,
			},
			changesetsCount: 2,
			campaignDiffStat: apitest.DiffStat{
				Added:   1 * changesetDiffStat.Added,
				Changed: 1 * changesetDiffStat.Changed,
				Deleted: 1 * changesetDiffStat.Deleted,
			},
		}
		testCampaignResponse(t, s, userCtx, input, want)

		for _, c := range changesets {
			// The changeset whose repository has been filtered should be hidden
			if c.RepoID == filteredRepo {
				testChangesetResponse(t, s, userCtx, c.ID, "HiddenExternalChangeset")
			} else {
				testChangesetResponse(t, s, userCtx, c.ID, "ExternalChangeset")
			}
		}

		// Now we query with more filters for the changesets. The hidden changesets
		// should not be returned, since that would leak information about the
		// hidden changesets.
		input = map[string]interface{}{
			"campaign":   string(campaigns.MarshalCampaignID(campaign.ID)),
			"checkState": string(campaigns.ChangesetCheckStatePassed),
		}
		wantCheckStateResponse := want
		wantCheckStateResponse.changesetsCount = 1
		wantCheckStateResponse.changesetTypes = map[string]int{
			"ExternalChangeset": 1,
			// No HiddenExternalChangeset
		}
		testCampaignResponse(t, s, userCtx, input, wantCheckStateResponse)

		input = map[string]interface{}{
			"campaign":    string(campaigns.MarshalCampaignID(campaign.ID)),
			"reviewState": string(campaigns.ChangesetReviewStateChangesRequested),
		}
		wantReviewStateResponse := want
		wantReviewStateResponse.changesetsCount = 1
		wantReviewStateResponse.changesetTypes = map[string]int{
			"ExternalChangeset": 1,
			// No HiddenExternalChangeset
		}
		testCampaignResponse(t, s, userCtx, input, wantReviewStateResponse)
	})

	t.Run("CampaignSpec and changesetSpecs", func(t *testing.T) {
		campaignSpec := &campaigns.CampaignSpec{
			UserID:          userID,
			NamespaceUserID: userID,
		}
		if err := store.CreateCampaignSpec(ctx, campaignSpec); err != nil {
			t.Fatal(err)
		}

		changesetSpecs := make([]*campaigns.ChangesetSpec, 0, len(repos))
		for _, r := range repos {
			c := &campaigns.ChangesetSpec{
				RepoID:         r.ID,
				UserID:         userID,
				CampaignSpecID: campaignSpec.ID,
			}
			if err := store.CreateChangesetSpec(ctx, c); err != nil {
				t.Fatal(err)
			}
			changesetSpecs = append(changesetSpecs, c)
		}

		// Query campaignSpec and check that we get all changesetSpecs
		userCtx := actor.WithActor(ctx, actor.FromUser(userID))
		testCampaignSpecResponse(t, s, userCtx, campaignSpec.RandID, wantCampaignSpecResponse{
			changesetSpecTypes:  map[string]int{"VisibleChangesetSpec": 2},
			changesetSpecsCount: 2,
		})

		// Now query the changesetSpecs as single nodes, to make sure that fetching/preloading
		// of repositories works
		for _, c := range changesetSpecs {
			// Both changesetSpecs are visible still, so both should be VisibleChangesetSpec
			testChangesetSpecResponse(t, s, userCtx, c.RandID, "VisibleChangesetSpec")
		}

		// Now we add the authzFilter and filter out the repository of one changeset
		filteredRepo := changesetSpecs[0].RepoID
		ct.AuthzFilterRepos(t, filteredRepo)

		// Send query again and check that for each filtered repository we get a
		// HiddenChangesetSpec.
		testCampaignSpecResponse(t, s, userCtx, campaignSpec.RandID, wantCampaignSpecResponse{
			changesetSpecTypes: map[string]int{
				"VisibleChangesetSpec": 1,
				"HiddenChangesetSpec":  1,
			},
			changesetSpecsCount: 2,
		})

		// Query the single changesetSpec nodes again
		for _, c := range changesetSpecs {
			// The changesetSpec whose repository has been filtered should be hidden
			if c.RepoID == filteredRepo {
				testChangesetSpecResponse(t, s, userCtx, c.RandID, "HiddenChangesetSpec")
			} else {
				testChangesetSpecResponse(t, s, userCtx, c.RandID, "VisibleChangesetSpec")
			}
		}
	})
}

type wantCampaignResponse struct {
	changesetTypes   map[string]int
	changesetsCount  int
	campaignDiffStat apitest.DiffStat
}

func testCampaignResponse(t *testing.T, s *graphql.Schema, ctx context.Context, in map[string]interface{}, w wantCampaignResponse) {
	t.Helper()

	var response struct{ Node apitest.Campaign }
	apitest.MustExec(ctx, t, s, in, &response, queryCampaignPermLevels)

	if have, want := response.Node.ID, in["campaign"]; have != want {
		t.Fatalf("campaign id is wrong. have %q, want %q", have, want)
	}

	if diff := cmp.Diff(w.changesetsCount, response.Node.Changesets.TotalCount); diff != "" {
		t.Fatalf("unexpected changesets total count (-want +got):\n%s", diff)
	}

	changesetTypes := map[string]int{}
	for _, c := range response.Node.Changesets.Nodes {
		changesetTypes[c.Typename]++
	}
	if diff := cmp.Diff(w.changesetTypes, changesetTypes); diff != "" {
		t.Fatalf("unexpected changesettypes (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(w.campaignDiffStat, response.Node.DiffStat); diff != "" {
		t.Fatalf("unexpected campaign diff stat (-want +got):\n%s", diff)
	}
}

const queryCampaignPermLevels = `
query($campaign: ID!, $reviewState: ChangesetReviewState, $checkState: ChangesetCheckState) {
  node(id: $campaign) {
    ... on Campaign {
      id

      changesets(first: 100, reviewState: $reviewState, checkState: $checkState) {
        totalCount
        nodes {
          __typename
          ... on HiddenExternalChangeset {
            id
          }
          ... on ExternalChangeset {
            id
            repository {
              id
              name
            }
          }
        }
      }

      diffStat {
        added
        changed
        deleted
      }
    }
  }
}
`

func testChangesetResponse(t *testing.T, s *graphql.Schema, ctx context.Context, id int64, wantType string) {
	t.Helper()

	var res struct{ Node apitest.Changeset }
	query := fmt.Sprintf(queryChangesetPermLevels, marshalChangesetID(id))
	apitest.MustExec(ctx, t, s, nil, &res, query)

	if have, want := res.Node.Typename, wantType; have != want {
		t.Fatalf("changeset has wrong typename. want=%q, have=%q", want, have)
	}

	if have, want := res.Node.ExternalState, string(campaigns.ChangesetExternalStateOpen); have != want {
		t.Fatalf("changeset has wrong state. want=%q, have=%q", want, have)
	}

	if have, want := res.Node.Campaigns.TotalCount, 1; have != want {
		t.Fatalf("changeset has wrong campaigns totalcount. want=%d, have=%d", want, have)
	}

	if parseJSONTime(t, res.Node.CreatedAt).IsZero() {
		t.Fatalf("changeset createdAt is zero")
	}

	if parseJSONTime(t, res.Node.UpdatedAt).IsZero() {
		t.Fatalf("changeset updatedAt is zero")
	}

	if parseJSONTime(t, res.Node.NextSyncAt).IsZero() {
		t.Fatalf("changeset next sync at is zero")
	}
}

const queryChangesetPermLevels = `
query {
  node(id: %q) {
    __typename

    ... on HiddenExternalChangeset {
      id

	  externalState
	  createdAt
	  updatedAt
	  nextSyncAt
	  campaigns {
	    totalCount
	  }
    }
    ... on ExternalChangeset {
      id

	  externalState
	  createdAt
	  updatedAt
	  nextSyncAt
	  campaigns {
	    totalCount
	  }

      repository {
        id
        name
      }
    }
  }
}
`

type wantCampaignSpecResponse struct {
	changesetSpecTypes  map[string]int
	changesetSpecsCount int
}

func testCampaignSpecResponse(t *testing.T, s *graphql.Schema, ctx context.Context, campaignSpecRandID string, w wantCampaignSpecResponse) {
	t.Helper()

	in := map[string]interface{}{
		"campaignSpec": string(marshalCampaignSpecRandID(campaignSpecRandID)),
	}

	var response struct{ Node apitest.CampaignSpec }
	apitest.MustExec(ctx, t, s, in, &response, queryCampaignSpecPermLevels)

	if have, want := response.Node.ID, in["campaignSpec"]; have != want {
		t.Fatalf("campaignSpec id is wrong. have %q, want %q", have, want)
	}

	if diff := cmp.Diff(w.changesetSpecsCount, response.Node.ChangesetSpecs.TotalCount); diff != "" {
		t.Fatalf("unexpected changesetSpecs total count (-want +got):\n%s", diff)
	}

	changesetSpecTypes := map[string]int{}
	for _, c := range response.Node.ChangesetSpecs.Nodes {
		changesetSpecTypes[c.Typename]++
	}
	if diff := cmp.Diff(w.changesetSpecTypes, changesetSpecTypes); diff != "" {
		t.Fatalf("unexpected changesetSpec types (-want +got):\n%s", diff)
	}
}

const queryCampaignSpecPermLevels = `
query($campaignSpec: ID!) {
  node(id: $campaignSpec) {
    ... on CampaignSpec {
      id

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
    }
  }
}
`

func testChangesetSpecResponse(t *testing.T, s *graphql.Schema, ctx context.Context, randID string, wantType string) {
	t.Helper()

	var res struct{ Node apitest.ChangesetSpec }
	query := fmt.Sprintf(queryChangesetSpecPermLevels, marshalChangesetSpecRandID(randID))
	apitest.MustExec(ctx, t, s, nil, &res, query)

	if have, want := res.Node.Typename, wantType; have != want {
		t.Fatalf("changesetspec has wrong typename. want=%q, have=%q", want, have)
	}
}

const queryChangesetSpecPermLevels = `
query {
  node(id: %q) {
    __typename

    ... on HiddenChangesetSpec {
      id
      type
    }

    ... on VisibleChangesetSpec {
      id
      type

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
`
