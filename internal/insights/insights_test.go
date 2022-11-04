package insights

import (
	"context"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold"
	"github.com/hexops/valast"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestMain(m *testing.M) {
	logtest.Init(m)
	os.Exit(m.Run())
}

func TestGetSearchInsights(t *testing.T) {
	ctx := context.Background()

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	_, err := db.ExecContext(ctx, `INSERT INTO orgs(id, name) VALUES (1, 'first-org'), (2, 'second-org');`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `

			INSERT INTO settings (id, org_id, contents, created_at, user_id, author_user_id)
			VALUES  (1, 1, $1, CURRENT_TIMESTAMP, NULL, NULL)`, insightSettingSimple)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetSearchInsights(ctx, db, All)
	if err != nil {
		t.Fatal(err)
	}

	weeks := 2

	want := []SearchInsight{{
		ID:           "searchInsights.insight.global.simple",
		Title:        "my insight",
		Repositories: []string{"github.com/sourcegraph/sourcegraph"},
		Series: []TimeSeries{{
			Name:   "Redis",
			Stroke: "var(--oc-red-7)",
			Query:  "redis",
		}},
		Step: Interval{
			Weeks: &weeks,
		},
	}}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatched search insight want/got: %v", diff)
	}
}

func TestGetSearchInsightsMulti(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	_, err := db.ExecContext(ctx, `INSERT INTO orgs(id, name) VALUES (1, 'first-org'), (2, 'second-org');`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `

			INSERT INTO settings (id, org_id, contents, created_at, user_id, author_user_id)
			VALUES  (1, 1, $1, CURRENT_TIMESTAMP, NULL, NULL)`, insightSettingNotSoSimple)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `

			INSERT INTO settings (id, org_id, contents, created_at, user_id, author_user_id)
			VALUES  (2, 2, $1, CURRENT_TIMESTAMP, NULL, NULL)`, insightSettingThree)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetSearchInsights(ctx, db, All)
	if err != nil {
		t.Fatal(err)
	}

	// sorting results for test determinism
	sort.Slice(got, func(i, j int) bool {
		return got[i].ID < got[j].ID
	})

	autogold.Want("testGetSearchInsightsMulti", []SearchInsight{
		{
			ID:           "searchInsights.insight.global.simple",
			Title:        "my insight",
			Repositories: []string{"github.com/sourcegraph/sourcegraph"},
			Series: []TimeSeries{{
				Name:   "Redis",
				Stroke: "var(--oc-red-7)",
				Query:  "redis",
			}},
			Step: Interval{Weeks: valast.Addr(2).(*int)},
		},
		{
			ID:           "searchInsights.insight.numbertwo",
			Title:        "numbertwo title",
			Repositories: []string{"github.com/sourcegraph/numbertwo"},
			Series: []TimeSeries{{
				Name:   "numbertwo series name",
				Stroke: "numbertwo var(--oc-red-7)",
				Query:  "numbertwo query",
			}},
			Step: Interval{Weeks: valast.Addr(2).(*int)},
		},
		{
			ID:           "searchInsights.insight.three",
			Title:        "three title",
			Repositories: []string{"github.com/sourcegraph/three"},
			Series: []TimeSeries{{
				Name:   "three series name",
				Stroke: "three var(--oc-red-7)",
				Query:  "three query",
			}},
			Step: Interval{Weeks: valast.Addr(4).(*int)},
		},
	}).Equal(t, got)
}

func TestGetIntegrationInsights(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	_, err := db.ExecContext(ctx, `INSERT INTO orgs(id, name) VALUES (1, 'first-org'), (2, 'second-org');`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `

			INSERT INTO settings (id, org_id, contents, created_at, user_id, author_user_id)
			VALUES  (1, 1, $1, CURRENT_TIMESTAMP, NULL, NULL)`, integratedInsightSimple)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetIntegratedInsights(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	// sorting for deterministic test
	sort.Slice(got, func(i, j int) bool {
		return strings.Compare(got[i].ID, got[j].ID) < 0
	})

	weeks := 2

	wantOrg := int32(1)
	want := []SearchInsight{
		{
			ID:           "unique-id1",
			Title:        "my insight",
			Repositories: []string{"github.com/sourcegraph/sourcegraph"},
			Series: []TimeSeries{{
				Name:   "Redis",
				Stroke: "var(--oc-red-7)",
				Query:  "redis",
			}},
			Step: Interval{
				Weeks: &weeks,
			},
			OrgID: &wantOrg,
		},
		{
			ID:           "unique-id2",
			Title:        "my insight2",
			Repositories: []string{"github.com/sourcegraph/sourcegraph"},
			Series: []TimeSeries{{
				Name:   "Redis",
				Stroke: "var(--oc-red-7)",
				Query:  "redis2",
			}},
			Step: Interval{
				Weeks: &weeks,
			},
			OrgID: &wantOrg,
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected integration insights want/got: %v", diff)
	}

	log15.Info("msg", "got", got)
}

const integratedInsightSimple = `{
	"insights.allrepos":{
		"unique-id1": {
			"title": "my insight",
			"repositories": ["github.com/sourcegraph/sourcegraph"],
			"series": [
			  {
				"name": "Redis",
				"query": "redis",
				"stroke": "var(--oc-red-7)"
			  }
			],
			"step": {
			  "weeks": 2
			}
		},
		"unique-id2": {
			"title": "my insight2",
			"repositories": ["github.com/sourcegraph/sourcegraph"],
			"series": [
			  {
				"name": "Redis",
				"query": "redis2",
				"stroke": "var(--oc-red-7)"
			  }
			],
			"step": {
			  "weeks": 2
			}
		},
		"invalid-schema": {
			"title": {}
		}
  	},
	"random-setting": {}
}`

const insightSettingSimple = `{"searchInsights.insight.global.simple": {
    "title": "my insight",
    "repositories": ["github.com/sourcegraph/sourcegraph"],
    "series": [
      {
        "name": "Redis",
        "query": "redis",
        "stroke": "var(--oc-red-7)"
      }
    ],
    "step": {
      "weeks": 2
    }
  }}`

const insightSettingThree = `{
	"searchInsights.insight.three": {
		"title": "three title",
		"repositories": ["github.com/sourcegraph/three"],
		"series": [
		  {
			"name": "three series name",
			"query": "three query",
			"stroke": "three var(--oc-red-7)"
		  }
		],
		"step": {
		  "weeks": 4
		}
	  }
}`

const insightSettingNotSoSimple = `{
  "searchInsights.insight.global.simple": {
    "title": "my insight",
    "repositories": ["github.com/sourcegraph/sourcegraph"],
    "series": [
      {
        "name": "Redis",
        "query": "redis",
        "stroke": "var(--oc-red-7)"
      }
    ],
    "step": {
      "weeks": 2
    }
  },
	"searchInsights.insight.numbertwo": {
		"title": "numbertwo title",
		"repositories": ["github.com/sourcegraph/numbertwo"],
		"series": [
		  {
			"name": "numbertwo series name",
			"query": "numbertwo query",
			"stroke": "numbertwo var(--oc-red-7)"
		  }
		],
		"step": {
		  "weeks": 2
		}
	  }
}`

func TestNextRecording(t *testing.T) {
	type args struct {
		current time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{name: "given first get next first", args: struct{ current time.Time }{current: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)}, want: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)},
		{name: "given december first get jan first", args: struct{ current time.Time }{current: time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)}, want: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		{name: "given december 31 get jan first", args: struct{ current time.Time }{current: time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)}, want: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NextRecording(tt.args.current); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NextRecording() = %v, want %v", got, tt.want)
			}
		})
	}
}
