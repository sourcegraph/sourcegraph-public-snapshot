package store

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/hexops/valast"

	"github.com/sourcegraph/log/logtest"

	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
)

func TestGetDashboard(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)

	_, err := insightsDB.ExecContext(context.Background(), `
		INSERT INTO dashboard (id, title)
		VALUES (1, 'test dashboard'), (2, 'private dashboard for user 3'), (3, 'private dashboard for org 1');`)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// assign some global grants just so the test can immediately fetch the created dashboard
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_grants (dashboard_id, global) VALUES (1, true)`)
	if err != nil {
		t.Fatal(err)
	}
	// assign a private user grant
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_grants (dashboard_id, user_id) VALUES (2, 3)`)
	if err != nil {
		t.Fatal(err)
	}

	// assign a private org grant
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_grants (dashboard_id, org_id) VALUES (3, 1)`)
	if err != nil {
		t.Fatal(err)
	}

	// create some views to assign to the dashboards
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO insight_view (id, title, description, unique_id)
									VALUES
										(1, 'my view', 'my description', 'unique1234'),
										(2, 'private view', 'private description', 'private1234'),
										(3, 'shared view', 'shared description', 'shared1234')`)
	if err != nil {
		t.Fatal(err)
	}

	// assign views to dashboards
	_, err = insightsDB.ExecContext(context.Background(), `INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id)
									VALUES
										(1, 1),
										(1, 3),
										(2, 2),
										(2, 3),
										(3, 3)`)
	if err != nil {
		t.Fatal(err)
	}

	store := NewDashboardStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test get all", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}

		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})

	t.Run("test user 3 can see global and user private dashboards", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{UserIDs: []int{3}})
		if err != nil {
			t.Fatal(err)
		}

		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
	t.Run("test user 3 can see both dashboards limit 1", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{UserIDs: []int{3}, Limit: 1})
		if err != nil {
			t.Fatal(err)
		}

		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
	t.Run("test user 3 can see both dashboards after 1", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{UserIDs: []int{3}, After: 1})
		if err != nil {
			t.Fatal(err)
		}

		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
	t.Run("test user 4 in org 1 can see both global and org private dashboard", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{UserIDs: []int{4}, OrgIDs: []int{1}})
		if err != nil {
			t.Fatal(err)
		}

		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
	t.Run("test user 3 can see both dashboards with view", func(t *testing.T) {
		viewId := "shared1234"
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{UserIDs: []int{3}, WithViewUniqueID: &viewId})
		if err != nil {
			t.Fatal(err)
		}

		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
	t.Run("test user 4 in org 1 can see both dashboards with view", func(t *testing.T) {
		viewId := "shared1234"
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{UserIDs: []int{4}, OrgIDs: []int{1}, WithViewUniqueID: &viewId})
		if err != nil {
			t.Fatal(err)
		}

		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
	t.Run("test user 4 can not see dashboards with private view", func(t *testing.T) {
		viewId := "private1234"
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{UserIDs: []int{4}, WithViewUniqueID: &viewId})
		if err != nil {
			t.Fatal(err)
		}

		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
	t.Run("test user 3 can see both dashboards with view limit 1", func(t *testing.T) {
		viewId := "shared1234"
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{UserIDs: []int{3}, WithViewUniqueID: &viewId, Limit: 1})
		if err != nil {
			t.Fatal(err)
		}

		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
	t.Run("test user 3 can see both dashboards with view after 1", func(t *testing.T) {
		viewId := "shared1234"
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{UserIDs: []int{3}, WithViewUniqueID: &viewId, After: 1})
		if err != nil {
			t.Fatal(err)
		}

		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
}

func TestCreateDashboard(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()
	store := NewDashboardStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test create dashboard", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*types.Dashboard{}).Equal(t, got)

		global := true
		orgId := 1
		grants := []DashboardGrant{{nil, nil, &global}, {nil, &orgId, nil}}
		_, err = store.CreateDashboard(ctx, CreateDashboardArgs{Dashboard: types.Dashboard{ID: 1, Title: "test dashboard 1"}, Grants: grants, UserIDs: []int{1}, OrgIDs: []int{1}})
		if err != nil {
			t.Fatal(err)
		}
		got, err = store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*types.Dashboard{{
			ID:           1,
			Title:        "test dashboard 1",
			UserIdGrants: []int64{},
			OrgIdGrants:  []int64{1},
			GlobalGrant:  true,
		}}).Equal(t, got)

		gotGrants, err := store.GetDashboardGrants(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*DashboardGrant{
			{
				Global: valast.Addr(true).(*bool),
			},
			{OrgID: valast.Addr(1).(*int)},
		}).Equal(t, gotGrants)
	})
}

func TestUpdateDashboard(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()
	store := NewDashboardStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	_, err := insightsDB.ExecContext(context.Background(), `
	INSERT INTO dashboard (id, title)
	VALUES (1, 'test dashboard 1'), (2, 'test dashboard 2');
	INSERT INTO dashboard_grants (dashboard_id, global)
	VALUES (1, true), (2, true);`)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("test update dashboard", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*types.Dashboard{
			{
				ID:           1,
				Title:        "test dashboard 1",
				UserIdGrants: []int64{},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
			{
				ID:           2,
				Title:        "test dashboard 2",
				UserIdGrants: []int64{},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
		}).Equal(t, got)

		newTitle := "new title!"
		global := true
		userId := 1
		grants := []DashboardGrant{{nil, nil, &global}, {&userId, nil, nil}}
		_, err = store.UpdateDashboard(ctx, UpdateDashboardArgs{1, &newTitle, grants, []int{1}, []int{}})
		if err != nil {
			t.Fatal(err)
		}
		got, err = store.GetDashboards(ctx, DashboardQueryArgs{UserIDs: []int{1}})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*types.Dashboard{
			{
				ID:           1,
				Title:        "new title!",
				UserIdGrants: []int64{1},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
			{
				ID:           2,
				Title:        "test dashboard 2",
				UserIdGrants: []int64{},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
		}).Equal(t, got)
	})
}

func TestDeleteDashboard(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	_, err := insightsDB.ExecContext(context.Background(), `
		INSERT INTO dashboard (id, title)
		VALUES (1, 'test dashboard 1'), (2, 'test dashboard 2');
		INSERT INTO dashboard_grants (dashboard_id, global)
		VALUES (1, true), (2, true);`)
	if err != nil {
		t.Fatal(err)
	}

	store := NewDashboardStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test delete dashboard", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*types.Dashboard{
			{
				ID:           1,
				Title:        "test dashboard 1",
				UserIdGrants: []int64{},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
			{
				ID:           2,
				Title:        "test dashboard 2",
				UserIdGrants: []int64{},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
		}).Equal(t, got)

		err = store.DeleteDashboard(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		got, err = store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*types.Dashboard{{
			ID:           2,
			Title:        "test dashboard 2",
			UserIdGrants: []int64{},
			OrgIdGrants:  []int64{},
			GlobalGrant:  true,
		}}).Equal(t, got)
	})
}

func TestRestoreDashboard(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	_, err := insightsDB.ExecContext(context.Background(), `
		INSERT INTO dashboard (id, title, deleted_at)
		VALUES (1, 'test dashboard 1', NULL), (2, 'test dashboard 2', NOW());
		INSERT INTO dashboard_grants (dashboard_id, global)
		VALUES (1, true), (2, true);`)
	if err != nil {
		t.Fatal(err)
	}

	store := NewDashboardStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test restore dashboard", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*types.Dashboard{
			{
				ID:           1,
				Title:        "test dashboard 1",
				UserIdGrants: []int64{},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
		}).Equal(t, got)

		err = store.RestoreDashboard(ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		got, err = store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*types.Dashboard{
			{
				ID:           1,
				Title:        "test dashboard 1",
				UserIdGrants: []int64{},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
			{
				ID:           2,
				Title:        "test dashboard 2",
				UserIdGrants: []int64{},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
		}).Equal(t, got)
	})
}

func TestAddViewsToDashboard(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	_, err := insightsDB.ExecContext(context.Background(), `
		INSERT INTO dashboard (id, title)
		VALUES (1, 'test dashboard 1'), (2, 'test dashboard 2');
		INSERT INTO dashboard_grants (dashboard_id, global)
		VALUES (1, true), (2, true);`)
	if err != nil {
		t.Fatal(err)
	}

	store := NewDashboardStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	t.Run("create and add view to dashboard", func(t *testing.T) {
		insightStore := NewInsightStore(insightsDB)
		view1, err := insightStore.CreateView(ctx, types.InsightView{
			Title:            "great view",
			Description:      "my view",
			UniqueID:         "view1234567",
			PresentationType: types.Line,
		}, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}
		view2, err := insightStore.CreateView(ctx, types.InsightView{
			Title:            "great view 2",
			Description:      "my view",
			UniqueID:         "view1234567-2",
			PresentationType: types.Line,
		}, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}

		dashboards, err := store.GetDashboards(ctx, DashboardQueryArgs{IDs: []int{1}})
		if err != nil || len(dashboards) != 1 {
			t.Errorf("failed to fetch dashboard before adding insight")
		}

		dashboard := dashboards[0]
		if len(dashboard.InsightIDs) != 0 {
			t.Errorf("unexpected value for insight views on dashboard before adding view")
		}
		err = store.AddViewsToDashboard(ctx, dashboard.ID, []string{view2.UniqueID, view1.UniqueID})
		if err != nil {
			t.Errorf("failed to add view to dashboard")
		}
		dashboards, err = store.GetDashboards(ctx, DashboardQueryArgs{IDs: []int{1}})
		if err != nil || len(dashboards) != 1 {
			t.Errorf("failed to fetch dashboard after adding insight")
		}
		got := dashboards[0]
		unsorted := got.InsightIDs
		sort.Slice(unsorted, func(i, j int) bool {
			return unsorted[i] < unsorted[j]
		})
		got.InsightIDs = unsorted
		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
}

func TestRemoveViewsFromDashboard(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	store := NewDashboardStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}

	insightStore := NewInsightStore(insightsDB)

	view, err := insightStore.CreateView(ctx, types.InsightView{
		Title:            "view1",
		Description:      "view1",
		UniqueID:         "view1",
		PresentationType: types.Line,
	}, []InsightViewGrant{GlobalGrant()})
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.CreateDashboard(ctx, CreateDashboardArgs{
		Dashboard: types.Dashboard{Title: "first", InsightIDs: []string{view.UniqueID}},
		Grants:    []DashboardGrant{GlobalDashboardGrant()},
		UserIDs:   []int{1},
		OrgIDs:    []int{1},
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.CreateDashboard(ctx, CreateDashboardArgs{
		Dashboard: types.Dashboard{Title: "second", InsightIDs: []string{view.UniqueID}},
		Grants:    []DashboardGrant{GlobalDashboardGrant()},
		UserIDs:   []int{1},
		OrgIDs:    []int{1},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("remove view from one dashboard only", func(t *testing.T) {
		dashboards, err := store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*types.Dashboard{
			{
				ID:           1,
				Title:        "first",
				InsightIDs:   []string{"view1"},
				UserIdGrants: []int64{},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
			{
				ID:           2,
				Title:        "second",
				InsightIDs:   []string{"view1"},
				UserIdGrants: []int64{},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
		}).Equal(t, dashboards)

		err = store.RemoveViewsFromDashboard(ctx, second.ID, []string{view.UniqueID})
		if err != nil {
			t.Fatal(err)
		}
		dashboards, err = store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Expect([]*types.Dashboard{
			{
				ID:           1,
				Title:        "first",
				InsightIDs:   []string{"view1"},
				UserIdGrants: []int64{},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
			{
				ID:           2,
				Title:        "second",
				UserIdGrants: []int64{},
				OrgIdGrants:  []int64{},
				GlobalGrant:  true,
			},
		}).Equal(t, dashboards)
	})
}

func TestHasDashboardPermission(t *testing.T) {
	logger := logtest.Scoped(t)
	insightsDB := edb.NewInsightsDB(dbtest.NewInsightsDB(logger, t), logger)
	now := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond).Round(0)
	ctx := context.Background()
	store := NewDashboardStore(insightsDB)
	store.Now = func() time.Time {
		return now
	}
	created, err := store.CreateDashboard(ctx, CreateDashboardArgs{
		Dashboard: types.Dashboard{
			Title: "test dashboard 123",
			Save:  true,
		},
		Grants:  []DashboardGrant{UserDashboardGrant(1), OrgDashboardGrant(5)},
		UserIDs: []int{1}, // this is a weird thing I'd love to get rid of, but for now this will cause the db to return
	})
	if err != nil {
		t.Fatal(err)
	}

	if created == nil {
		t.Fatalf("nil dashboard")
	}

	second, err := store.CreateDashboard(ctx, CreateDashboardArgs{
		Dashboard: types.Dashboard{
			Title: "second test dashboard",
			Save:  true,
		},
		Grants:  []DashboardGrant{UserDashboardGrant(2), OrgDashboardGrant(5)},
		UserIDs: []int{2}, // this is a weird thing I'd love to get rid of, but for now this will cause the db to return
	})
	if err != nil {
		t.Fatal(err)
	}

	if second == nil {
		t.Fatalf("nil dashboard")
	}

	tests := []struct {
		name                 string
		shouldHavePermission bool
		userIds              []int
		orgIds               []int
		dashboardIDs         []int
	}{
		{
			name:                 "user 1 has access to dashboard",
			shouldHavePermission: true,
			userIds:              []int{1},
			orgIds:               nil,
			dashboardIDs:         []int{created.ID},
		},
		{
			name:                 "user 3 does not have access to dashboard",
			shouldHavePermission: false,
			userIds:              []int{3},
			orgIds:               nil,
			dashboardIDs:         []int{created.ID},
		},
		{
			name:                 "org 5 has access to dashboard",
			shouldHavePermission: true,
			userIds:              nil,
			orgIds:               []int{5},
			dashboardIDs:         []int{created.ID},
		},
		{
			name:                 "org 7 does not have access to dashboard",
			shouldHavePermission: false,
			userIds:              nil,
			orgIds:               []int{7},
			dashboardIDs:         []int{created.ID},
		},
		{
			name:                 "no access when dashboard does not exist",
			shouldHavePermission: false,
			userIds:              []int{3},
			orgIds:               []int{5},
			dashboardIDs:         []int{-2},
		},
		{
			name:                 "user 1 has access to one of two dashboards",
			shouldHavePermission: false,
			userIds:              []int{1},
			orgIds:               nil,
			dashboardIDs:         []int{created.ID, second.ID},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := store.HasDashboardPermission(ctx, test.dashboardIDs, test.userIds, test.orgIds)
			if err != nil {
				t.Error(err)
			}
			want := test.shouldHavePermission
			if want != got {
				t.Errorf("unexpected dashboard access result from HasDashboardPermission: want: %v got: %v", want, got)
			}
		})
	}
}
