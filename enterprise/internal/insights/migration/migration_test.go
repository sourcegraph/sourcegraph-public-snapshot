package migration

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	insightsdbtesting "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/dbtesting"
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
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()

	migrator := migrator{insightStore: store.NewInsightStore(timescale)}

	t.Run("match on org ID", func(t *testing.T) {
		want := "myInsight-org-3"

		migration := migrationContext{
			userId: 0,
			orgIds: []int{3},
		}

		_, err := timescale.Exec("insert into insight_view (unique_id) values ($1);", want)
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

		_, err := timescale.Exec("insert into insight_view (unique_id) values ($1);", want)
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

		_, err := timescale.Exec("insert into insight_view (unique_id) values ($1);", want)
		if err != nil {
			t.Fatal(err)
		}
		// this one should NOT match
		_, err = timescale.Exec("insert into insight_view (unique_id) values ($1);", "myInsight3-org-5")
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
