package resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	internalTypes "github.com/sourcegraph/sourcegraph/internal/types"
)

// Note: You can `go test ./resolvers -update` to update the expected `want` values in these tests.
// See https://github.com/hexops/autogold for more information.

// TestResolver_InsightConnection tests that the InsightConnection GraphQL resolver works.
func TestResolver_InsightConnection(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(t))

	testSetup := func(t *testing.T) (context.Context, graphqlbackend.InsightConnectionResolver) {
		// Setup the GraphQL resolver.
		ctx := actor.WithInternalActor(context.Background())
		now := time.Now().UTC().Truncate(time.Microsecond)
		clock := func() time.Time { return now }

		postgres := database.NewDB(dbtest.NewDB(t))
		resolver := newWithClock(insightsDB, postgres, clock)

		insightMetadataStore := store.NewMockInsightMetadataStore()
		insightMetadataStore.GetMappedFunc.SetDefaultReturn([]types.Insight{
			{
				UniqueID:    "unique1",
				Title:       "title1",
				Description: "desc1",
				Series: []types.InsightViewSeries{
					{
						UniqueID:           "unique1",
						SeriesID:           "1234567",
						Title:              "title1",
						Description:        "desc1",
						Query:              "query1",
						CreatedAt:          now,
						OldestHistoricalAt: now,
						LastRecordedAt:     now,
						NextRecordingAfter: now,
						Label:              "label1",
						LineColor:          "color1",
					},
				},
			},
		}, nil)
		resolver.insightMetadataStore = insightMetadataStore

		// Create the insights connection resolver.
		conn, err := resolver.Insights(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}

		return ctx, conn
	}

	t.Run("TotalCount", func(t *testing.T) {
		ctx, conn := testSetup(t)
		got, err := conn.TotalCount(ctx)
		if err != nil {
			t.Fatal(err)
		}
		want := int32(1)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("incorrect number of insights (want/got): %v", diff)
		}
	})

	t.Run("PageInfo", func(t *testing.T) {
		// TODO: future: our pagination support is non-existent. Currently we just return all
		// insights, regardless of how many you ask for.
		ctx, conn := testSetup(t)
		gotPageInfo, err := conn.PageInfo(ctx)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("PageInfo", &graphqlutil.PageInfo{}).Equal(t, gotPageInfo)
	})

	t.Run("Nodes", func(t *testing.T) {
		ctx, conn := testSetup(t)
		nodes, err := conn.Nodes(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(nodes) != 1 {
			t.Fatal("incorrect length")
		}
		autogold.Want("first insight", map[string]any{"description": "desc1", "title": "title1"}).Equal(t, map[string]any{
			"title":       nodes[0].Title(),
			"description": nodes[0].Description(),
		})
		// TODO(slimsag): put series length into map (autogold bug, omits the field for some reason?)
		autogold.Want("first insight: series length", int(1)).Equal(t, len(nodes[0].Series()))
	})
}

func TestResolver_InsightsRepoPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(t))
	postgres := database.NewDB(dbtest.NewDB(t))

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time { return now }
	authz.SetProviders(false, []authz.Provider{}) // setting authz in this way will force user permissions to be enabled

	// Set up an external service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	externalService := &internalTypes.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GITHUB #1",
		Config:      `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc", "authorization": {}}`,
	}
	err := postgres.ExternalServices().Create(context.Background(), confGet, externalService)
	if err != nil {
		t.Fatal(err)
	}

	// Create three repositories -
	// 1) private repo that will be assigned to user 1
	// 2 & 3) public repos
	_, err = postgres.Handle().DBUtilDB().ExecContext(ctx, `
		INSERT INTO repo (id, name, description, fork, created_at, updated_at, external_id, external_service_type,
					  external_service_id, archived, uri, deleted_at, metadata, private, stars)
		VALUES
			(1, 'test-repo1', 'description', false, current_timestamp, current_timestamp, 1, 'github', 1, false, 'github.com/test-repo/test-repo1', null, '{}', true, 1),
			(2, 'test-repo2', 'description', false, current_timestamp, current_timestamp, 2, 'github', 1, false, 'github.com/test-repo/test-repo2', null, '{}', false, 1),
			(3, 'test-repo3', 'description', false, current_timestamp, current_timestamp, 3, 'github', 1, false, 'github.com/test-repo/test-repo3', null, '{}', false, 1);

		INSERT INTO user_permissions (user_id, permission, object_type, updated_at, synced_at, object_ids_ints)
		VALUES
		       (1, 'read', 'repos', current_timestamp, current_timestamp, ARRAY[1]);

		INSERT INTO user_permissions (user_id, permission, object_type, updated_at, synced_at)
		VALUES
		       (2, 'read', 'repos', current_timestamp, current_timestamp),
		       (3, 'read', 'repos', current_timestamp, current_timestamp);

		INSERT INTO users (id, username, display_name, avatar_url, created_at, updated_at, deleted_at, invite_quota, passwd,
						   site_admin)
		VALUES
		(1, 'user1', 'user1', null, current_timestamp, current_timestamp, null, 1, 'abc', false),
		(2, 'user2', 'user2', null, current_timestamp, current_timestamp, null, 1, 'abc', false);`,
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = postgres.Handle().DBUtilDB().ExecContext(ctx, `
		INSERT INTO external_service_repos (external_service_id, repo_id, clone_url)
		VALUES
		       ($1, 1, ''),
		       ($1, 2, ''),
		       ($1, 3, '');
`, externalService.ID,
	)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO repo_names (name) VALUES ($1);`, fmt.Sprint("ignore-me-", i))
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create some timeseries data, one row in each repository
	_, err = insightsDB.ExecContext(context.Background(), `
		INSERT INTO series_points (series_id, "time", "value", metadata_id, repo_id, repo_name_id, original_repo_name_id)
		VALUES
			('s:087855E6A24440837303FD8A252E9893E8ABDFECA55B61AC83DA1B521906626E', $1, 5.0, null, 1, 3, 3),
			('s:087855E6A24440837303FD8A252E9893E8ABDFECA55B61AC83DA1B521906626E', $1, 6.0, null, 2, 2, 2),
			('s:087855E6A24440837303FD8A252E9893E8ABDFECA55B61AC83DA1B521906626E', $1, 7.0, null, 3, 1, 1)`, now)
	if err != nil {
		t.Fatal(err)
	}

	setUpTest := func(ctx context.Context, t *testing.T) graphqlbackend.InsightConnectionResolver {

		resolver := newWithClock(insightsDB, postgres, clock)
		insightMetadataStore := store.NewMockInsightMetadataStore()
		insightMetadataStore.GetMappedFunc.SetDefaultReturn([]types.Insight{
			{
				UniqueID:    "unique1",
				Title:       "title1",
				Description: "desc1",
				Series: []types.InsightViewSeries{
					{
						UniqueID:           "unique1",
						SeriesID:           "s:087855E6A24440837303FD8A252E9893E8ABDFECA55B61AC83DA1B521906626E",
						Title:              "title1",
						Description:        "desc1",
						Query:              "query1",
						CreatedAt:          now,
						OldestHistoricalAt: now,
						LastRecordedAt:     now,
						NextRecordingAfter: now,
						Label:              "label1",
						LineColor:          "color1",
					},
				},
			},
		}, nil)
		resolver.insightMetadataStore = insightMetadataStore
		conn, err := resolver.Insights(ctx, &graphqlbackend.InsightsArgs{})
		if err != nil {
			t.Fatal(err)
		}

		return conn
	}

	t.Run("user with private repo", func(t *testing.T) {
		ctx := context.Background()
		a := actor.FromUser(1)
		ctx = actor.WithActor(ctx, a)
		conn := setUpTest(ctx, t)

		want := []point{
			{
				value: 18.0,
				time:  now,
			},
		}

		got := resolvePoints(ctx, conn, t)
		if diff := cmp.Diff(want, got, cmp.Comparer(func(x, y point) bool {
			return x.value == y.value && x.time.Equal(y.time)
		})); diff != "" {
			t.Errorf("expected results for user with private repo diff: %v", diff)
		}
	})

	t.Run("user with only public repos", func(t *testing.T) {
		ctx := context.Background()
		a := actor.FromUser(2)
		ctx = actor.WithActor(ctx, a)
		conn := setUpTest(ctx, t)

		want := []point{
			{
				value: 13.0,
				time:  now,
			},
		}

		got := resolvePoints(ctx, conn, t)
		if diff := cmp.Diff(want, got, cmp.Comparer(func(x, y point) bool {
			return x.value == y.value && x.time.Equal(y.time)
		})); diff != "" {
			t.Errorf("expected results for user with only public repos diff: %v", diff)
		}
	})

}

type point struct {
	value float64
	time  time.Time
}

func resolvePoints(ctx context.Context, conn graphqlbackend.InsightConnectionResolver, t *testing.T) []point {
	nodes, err := conn.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) < 1 {
		t.Fatalf("unexpected length of nodes: want: %v got: %v", 1, len(nodes))
	}

	expected := nodes[0]
	seriesResolvers := expected.Series()
	if len(seriesResolvers) != 1 {
		t.Errorf("unexpected length of series resolvers: want: %v got: %v", 1, len(seriesResolvers))
	}
	sr := seriesResolvers[0]
	data, err := sr.Points(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	var results []point
	for _, d := range data {
		var temp point
		temp.value = d.Value()
		temp.time = d.DateTime().Time
		results = append(results, temp)
	}
	return results
}
