package store

import (
	"context"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	insightsdbtesting "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/dbtesting"
)

func TestGet(t *testing.T) {
	timescale, cleanup := insightsdbtesting.TimescaleDB(t)
	defer cleanup()

	_, err := timescale.Exec(`INSERT INTO insight_view (title, description, unique_id)
									VALUES ('test title', 'test description', 'unique-1'),
									       ('test title 2', 'description2', 'unique-2');`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = timescale.Exec(`INSERT INTO insight_series (series_id, query, recording_interval_days)
									VALUES ('series-id-1', 'query-1', 5), ('series-id-2', 'query-2', 6);`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = timescale.Exec(`INSERT INTO insight_view_series (insight_view_id, insight_series_id, label, stroke)
									VALUES (1, 1, 'label1', 'color1'), (1, 2, 'label2', 'color2'), (2, 2, 'second-label-2', 'second-color-2');`)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	t.Run("test get all", func(t *testing.T) {
		store := NewInsightStore(timescale)

		got, err := store.Get(ctx, InsightQueryArgs{})
		if err != nil {
			t.Fatal(err)
		}

		var want []types.InsightViewSeries

		autogold.Want("want-all-results", want).Equal(t, got)
	})
}
