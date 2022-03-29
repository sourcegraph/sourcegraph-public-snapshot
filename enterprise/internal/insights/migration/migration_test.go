package migration

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"

	"github.com/hexops/autogold"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
)

func TestBuildUniqueIdCondition(t *testing.T) {
	insightId := "myInsight"
	tests := []struct {
		name             string
		migrationContext migrationContext
		result           string
	}{
		{name: "only orgs", migrationContext: migrationContext{
			userId: 0,
			orgIds: []int{5, 6, 7},
		}, result: "myInsight-%(org-5|org-6|org-7)%"},
		{name: "only user", migrationContext: migrationContext{
			userId: 1,
		}, result: "myInsight-%(user-1)%"},
		{name: "both user and orgs", migrationContext: migrationContext{
			userId: 7,
			orgIds: []int{4, 5},
		}, result: "myInsight-%(org-4|org-5|user-7)%"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if diff := cmp.Diff(test.result, test.migrationContext.buildUniqueIdCondition(insightId)); diff != "" {
				t.Errorf("mismatched insight id condition (want/got): %s", diff)
			}
		})
	}
}

func TestToInsightUniqueIdQuery(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	insightsDB := dbtest.NewInsightsDB(t)

	migrator := migrator{insightStore: store.NewInsightStore(insightsDB)}

	t.Run("match on org ID", func(t *testing.T) {
		want := "myInsight-org-3"

		migration := migrationContext{
			userId: 0,
			orgIds: []int{3},
		}

		_, err := insightsDB.Exec("insert into insight_view (unique_id) values ($1);", want)
		if err != nil {
			t.Fatal(err)
		}

		got, found, err := migrator.lookupUniqueId(ctx, migration, "myInsight")
		if err != nil {
			t.Fatal(err)
		} else if !found {
			t.Fatal("insight not found")
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatched unique insight id (want/got): %s", diff)
		}
	})

	t.Run("match on user ID", func(t *testing.T) {
		want := "myInsight2-user-1"

		migration := migrationContext{
			userId: 1,
		}

		_, err := insightsDB.Exec("insert into insight_view (unique_id) values ($1);", want)
		if err != nil {
			t.Fatal(err)
		}

		got, found, err := migrator.lookupUniqueId(ctx, migration, "myInsight2")
		if err != nil {
			t.Fatal(err)
		} else if !found {
			t.Fatal("insight not found")
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatched unique insight id (want/got): %s", diff)
		}
	})

	t.Run("match on specific Org", func(t *testing.T) {
		want := "myInsight3-org-3"

		migration := migrationContext{
			orgIds: []int{3},
		}

		_, err := insightsDB.Exec("insert into insight_view (unique_id) values ($1);", want)
		if err != nil {
			t.Fatal(err)
		}
		// this one should NOT match
		_, err = insightsDB.Exec("insert into insight_view (unique_id) values ($1);", "myInsight3-org-5")
		if err != nil {
			t.Fatal(err)
		}

		got, found, err := migrator.lookupUniqueId(ctx, migration, "myInsight3")
		if err != nil {
			t.Fatal(err)
		} else if !found {
			t.Fatal("insight not found")
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatched unique insight id (want/got): %s", diff)
		}
	})
}

func TestSpecialCaseDashboardTitle(t *testing.T) {
	t.Run("global title", func(t *testing.T) {
		want := "Global Insights"
		got := specialCaseDashboardTitle("Global")
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected special case dashboard title (want/got): %v", diff)
		}
	})
	t.Run("non global title", func(t *testing.T) {
		want := "First Last's Insights"
		got := specialCaseDashboardTitle("First Last")
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected special case dashboard title (want/got): %v", diff)
		}
	})
}

func TestCreateSpecialCaseDashboard(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	insightsDB := dbtest.NewInsightsDB(t)
	migrator := migrator{insightStore: store.NewInsightStore(insightsDB), dashboardStore: store.NewDashboardStore(insightsDB)}

	newView := func(insightId string) {
		_, err := insightsDB.Exec("insert into insight_view (unique_id) values ($1);", insightId)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("user special dashboard", func(t *testing.T) {
		subjectName := "Samwise Gamgee"
		insightReferences := []string{"ringsThrownIntoMountDoom1", "wateringTheGarden1", "hobbitsInTheShire1"}
		for _, reference := range insightReferences {
			newView(reference + "-user-1")
		}
		migrationContext := migrationContext{userId: 1}
		got, err := migrator.createSpecialCaseDashboard(ctx, subjectName, insightReferences, migrationContext)
		if err != nil {
			t.Error(err)
		}
		// setting ID specifically for test determinism
		got.ID = 1
		autogold.Want("user special dashboard", types.Dashboard{
			ID: 1, Title: "Samwise Gamgee's Insights",
			InsightIDs: []string{
				"ringsThrownIntoMountDoom1-user-1",
				"wateringTheGarden1-user-1",
				"hobbitsInTheShire1-user-1",
			},
			UserIdGrants: []int64{1},
			OrgIdGrants:  []int64{},
		}).Equal(t, *got)
	})

	t.Run("user special dashboard with pretransformed insight Ids", func(t *testing.T) {
		subjectName := "Samwise Gamgee"
		insightReferences := []string{"ringsThrownIntoMountDoom2-user-1", "wateringTheGarden2-org-5", "hobbitsInTheShire2-org-6"}
		for _, reference := range insightReferences {
			newView(reference)
		}
		migrationContext := migrationContext{userId: 1, orgIds: []int{5, 6}}
		got, err := migrator.createSpecialCaseDashboard(ctx, subjectName, insightReferences, migrationContext)
		if err != nil {
			t.Error(err)
		}
		// setting ID specifically for test determinism
		got.ID = 1
		autogold.Want("user special dashboard with pretransformed insight Ids", types.Dashboard{
			ID: 1, Title: "Samwise Gamgee's Insights",
			InsightIDs: []string{
				"ringsThrownIntoMountDoom2-user-1",
				"wateringTheGarden2-org-5",
				"hobbitsInTheShire2-org-6",
			},
			UserIdGrants: []int64{1},
			OrgIdGrants:  []int64{},
		}).Equal(t, *got)
	})
	t.Run("org special dashboard", func(t *testing.T) {
		subjectName := "The Shire"
		insightReferences := []string{"ringsThrownIntoMountDoom3", "wateringTheGarden3", "hobbitsInTheShire3"}
		for _, reference := range insightReferences {
			newView(reference + "-org-1")
		}
		migrationContext := migrationContext{orgIds: []int{1}}
		got, err := migrator.createSpecialCaseDashboard(ctx, subjectName, insightReferences, migrationContext)
		if err != nil {
			t.Error(err)
		}
		// setting ID specifically for test determinism
		got.ID = 1
		autogold.Want("org special dashboard", types.Dashboard{
			ID: 1, Title: "The Shire's Insights",
			InsightIDs: []string{
				"ringsThrownIntoMountDoom3-org-1",
				"wateringTheGarden3-org-1",
				"hobbitsInTheShire3-org-1",
			},
			UserIdGrants: []int64{},
			OrgIdGrants:  []int64{1},
		}).Equal(t, *got)
	})
	t.Run("global special dashboard", func(t *testing.T) {
		subjectName := "Global"
		insightReferences := []string{"istariInMiddleEarth"}
		for _, reference := range insightReferences {
			newView(reference)
		}
		migrationContext := migrationContext{}
		got, err := migrator.createSpecialCaseDashboard(ctx, subjectName, insightReferences, migrationContext)
		if err != nil {
			t.Error(err)
		}
		// setting ID specifically for test determinism
		got.ID = 1
		autogold.Want("global special dashboard", types.Dashboard{
			ID: 1, Title: "Global Insights", InsightIDs: []string{
				"istariInMiddleEarth",
			},
			UserIdGrants: []int64{},
			OrgIdGrants:  []int64{},
			GlobalGrant:  true,
		}).Equal(t, *got)
	})
	t.Run("global special dashboard with no insights", func(t *testing.T) {
		subjectName := "Global"
		var insightReferences []string
		migrationContext := migrationContext{}
		got, err := migrator.createSpecialCaseDashboard(ctx, subjectName, insightReferences, migrationContext)
		if err != nil {
			t.Error(err)
		}
		// setting ID specifically for test determinism
		got.ID = 1
		autogold.Want("global special dashboard with no insights", types.Dashboard{
			ID: 1, Title: "Global Insights", UserIdGrants: []int64{},
			OrgIdGrants: []int64{},
			GlobalGrant: true,
		}).Equal(t, *got)
	})
}
