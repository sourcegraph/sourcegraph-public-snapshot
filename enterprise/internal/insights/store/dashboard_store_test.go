package store

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold"

	insightsdbtesting "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/dbtesting"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
)

func TestGetDashboard(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Truncate(time.Microsecond).Round(0)

	_, err := timescale.Exec(`
		INSERT INTO dashboard (id, title)
		VALUES (1, 'test dashboard'), (2, 'private dashboard for user 3');`)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// assign some global grants just so the test can immediately fetch the created dashboard
	_, err = timescale.Exec(`INSERT INTO dashboard_grants (dashboard_id, global)
									VALUES (1, true)`)
	if err != nil {
		t.Fatal(err)
	}
	// assign a private grant
	_, err = timescale.Exec(`INSERT INTO dashboard_grants (dashboard_id, user_id)
									VALUES (2, 3)`)
	if err != nil {
		t.Fatal(err)
	}

	// assign some global grants just so the test can immediately fetch the created dashboard
	_, err = timescale.Exec(`INSERT INTO insight_view (id, title, description, unique_id)
									VALUES (1, 'my view', 'my description', 'unique1234')`)
	if err != nil {
		t.Fatal(err)
	}

	// assign some global grants just so the test can immediately fetch the created dashboard
	_, err = timescale.Exec(`INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id)
									VALUES (1, 1)`)
	if err != nil {
		t.Fatal(err)
	}

	store := NewDashboardStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test get all", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}

		autogold.Equal(t, got, autogold.ExportedOnly())
	})

	t.Run("test user 3 can see both dashboards", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{UserID: []int{3}})
		if err != nil {
			t.Fatal(err)
		}

		autogold.Equal(t, got, autogold.ExportedOnly())
	})
	t.Run("test user 3 can see both dashboards limit 1", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{UserID: []int{3}, Limit: 1})
		if err != nil {
			t.Fatal(err)
		}

		autogold.Equal(t, got, autogold.ExportedOnly())
	})
	t.Run("test user 3 can see both dashboards after 1", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{UserID: []int{3}, After: 1})
		if err != nil {
			t.Fatal(err)
		}

		autogold.Equal(t, got, autogold.ExportedOnly())
	})
}

func TestDeleteDashboard(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	_, err := timescale.Exec(`
		INSERT INTO dashboard (id, title)
		VALUES (1, 'test dashboard 1'), (2, 'test dashboard 2');
		INSERT INTO dashboard_grants (dashboard_id, global)
		VALUES (1, true), (2, true);`)
	if err != nil {
		t.Fatal(err)
	}

	store := NewDashboardStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("test delete dashboard", func(t *testing.T) {
		got, err := store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("BeforeDelete", []*types.Dashboard{
			{
				ID:    1,
				Title: "test dashboard 1",
			},
			{
				ID:    2,
				Title: "test dashboard 2",
			}}).Equal(t, got)

		err = store.DeleteDashboard(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		got, err = store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("AfterDelete", []*types.Dashboard{{
			ID:    2,
			Title: "test dashboard 2",
		}}).Equal(t, got)
	})
}

func TestAssociateViewsById(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	_, err := timescale.Exec(`
		INSERT INTO dashboard (id, title)
		VALUES (1, 'test dashboard 1'), (2, 'test dashboard 2');
		INSERT INTO dashboard_grants (dashboard_id, global)
		VALUES (1, true), (2, true);`)
	if err != nil {
		t.Fatal(err)
	}

	store := NewDashboardStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	t.Run("create and add view to dashboard", func(t *testing.T) {
		insightStore := NewInsightStore(timescale)
		view, err := insightStore.CreateView(ctx, types.InsightView{
			Title:       "great view",
			Description: "my view",
			UniqueID:    "view1234567",
		}, []InsightViewGrant{GlobalGrant()})
		if err != nil {
			t.Fatal(err)
		}

		dashboards, err := store.GetDashboards(ctx, DashboardQueryArgs{ID: 1})
		if err != nil || len(dashboards) != 1 {
			t.Errorf("failed to fetch dashboard before adding insight")
		}

		dashboard := dashboards[0]
		if len(dashboard.InsightIDs) != 0 {
			t.Errorf("unexpected value for insight views on dashboard before adding view")
		}
		err = store.AddViewsToDashboard(ctx, dashboard.ID, []string{view.UniqueID})
		if err != nil {
			t.Errorf("failed to add view to dashboard")
		}
		dashboards, err = store.GetDashboards(ctx, DashboardQueryArgs{ID: 1})
		if err != nil || len(dashboards) != 1 {
			t.Errorf("failed to fetch dashboard after adding insight")
		}
		got := dashboards[0]
		autogold.Want("check views are added to dashboard", &types.Dashboard{ID: 1, Title: "test dashboard 1", InsightIDs: []string{
			"view1234567",
		}}).Equal(t, got)
	})
}

func TestRemoveViewsFromDashboard(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()
	now := time.Now().Truncate(time.Microsecond).Round(0)
	ctx := context.Background()

	store := NewDashboardStore(timescale)
	store.Now = func() time.Time {
		return now
	}

	insightStore := NewInsightStore(timescale)

	view, err := insightStore.CreateView(ctx, types.InsightView{
		Title:       "view1",
		Description: "view1",
		UniqueID:    "view1",
	}, []InsightViewGrant{GlobalGrant()})
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.CreateDashboard(ctx, types.Dashboard{Title: "first", InsightIDs: []string{view.UniqueID}}, []DashboardGrant{GlobalDashboardGrant()})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.CreateDashboard(ctx, types.Dashboard{Title: "second", InsightIDs: []string{view.UniqueID}}, []DashboardGrant{GlobalDashboardGrant()})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("remove view from one dashboard only", func(t *testing.T) {
		dashboards, err := store.GetDashboards(ctx, DashboardQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}
		autogold.Want("dashboards before removing a view", []*types.Dashboard{
			{
				ID:         1,
				Title:      "first",
				InsightIDs: []string{"view1"},
			},
			{
				ID:         2,
				Title:      "second",
				InsightIDs: []string{"view1"},
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
		autogold.Want("dashboards after removing a view", []*types.Dashboard{
			{
				ID:         1,
				Title:      "first",
				InsightIDs: []string{"view1"},
			},
			{
				ID:    2,
				Title: "second",
			},
		}).Equal(t, dashboards)
	})

}
