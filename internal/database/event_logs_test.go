package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

func TestEventLogs_ValidInfo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()

	var testCases = []struct {
		name  string
		event *Event
		err   string // Stringified error
	}{
		{
			name:  "EmptyName",
			event: &Event{UserID: 1, URL: "http://sourcegraph.com", Source: "WEB"},
			err:   `INSERT: ERROR: new row for relation "event_logs" violates check constraint "event_logs_check_name_not_empty" (SQLSTATE 23514)`,
		},
		{
			name:  "InvalidUser",
			event: &Event{Name: "test_event", URL: "http://sourcegraph.com", Source: "WEB"},
			err:   `INSERT: ERROR: new row for relation "event_logs" violates check constraint "event_logs_check_has_user" (SQLSTATE 23514)`,
		},
		{
			name:  "EmptySource",
			event: &Event{Name: "test_event", URL: "http://sourcegraph.com", UserID: 1},
			err:   `INSERT: ERROR: new row for relation "event_logs" violates check constraint "event_logs_check_source_not_empty" (SQLSTATE 23514)`,
		},
		{
			name:  "ValidInsert",
			event: &Event{Name: "test_event", UserID: 1, URL: "http://sourcegraph.com", Source: "WEB"},
			err:   "<nil>",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := EventLogs(db).Insert(ctx, tc.event)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("have %+v, want %+v", have, want)
			}
		})
	}
}

func TestEventLogs_CountUniqueUsersPerPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()

	now := time.Now()
	startDate, _ := calcStartDate(now, Daily, 3)
	secondDay := startDate.Add(time.Hour * 24)
	thirdDay := startDate.Add(time.Hour * 24 * 2)

	events := []*Event{
		makeTestEvent(&Event{UserID: 1, Timestamp: startDate}),
		makeTestEvent(&Event{UserID: 1, Timestamp: startDate}),
		makeTestEvent(&Event{UserID: 2, Timestamp: startDate}),
		makeTestEvent(&Event{UserID: 2, Timestamp: startDate}),

		makeTestEvent(&Event{UserID: 1, Timestamp: secondDay}),
		makeTestEvent(&Event{UserID: 2, Timestamp: secondDay}),
		makeTestEvent(&Event{UserID: 3, Timestamp: secondDay}),
		makeTestEvent(&Event{UserID: 1, Timestamp: secondDay}),

		makeTestEvent(&Event{UserID: 5, Timestamp: thirdDay}),
		makeTestEvent(&Event{UserID: 6, Timestamp: thirdDay}),
		makeTestEvent(&Event{UserID: 7, Timestamp: thirdDay}),
		makeTestEvent(&Event{UserID: 8, Timestamp: thirdDay}),
	}

	for _, e := range events {
		if err := EventLogs(db).Insert(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	values, err := EventLogs(db).CountUniqueUsersPerPeriod(ctx, Daily, now, 3, nil)
	if err != nil {
		t.Fatal(err)
	}

	assertUsageValue(t, values[0], startDate.Add(time.Hour*24*2), 4)
	assertUsageValue(t, values[1], startDate.Add(time.Hour*24), 3)
	assertUsageValue(t, values[2], startDate, 2)
}

func TestEventLogs_UsersUsageCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()

	now := time.Now()

	startDate, _ := calcStartDate(now, Daily, 3)
	secondDay := startDate.Add(time.Hour * 24)
	thirdDay := startDate.Add(time.Hour * 24 * 2)

	days := []time.Time{startDate, secondDay, thirdDay}
	names := []string{"SearchResultsQueried", "codeintel"}
	users := []uint32{1, 2}

	for _, day := range days {
		for _, user := range users {
			for _, name := range names {
				for i := 0; i < 25; i++ {
					e := &Event{
						UserID:    user,
						Name:      name,
						URL:       "test",
						Source:    "test",
						Timestamp: day.Add(time.Minute * time.Duration(rand.Intn(60*12))),
					}

					if err := EventLogs(db).Insert(ctx, e); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}

	have, err := EventLogs(db).UsersUsageCounts(ctx)
	if err != nil {
		t.Fatal(err)
	}

	want := []types.UserUsageCounts{
		{Date: days[2], UserID: users[0], SearchCount: 25, CodeIntelCount: 25},
		{Date: days[2], UserID: users[1], SearchCount: 25, CodeIntelCount: 25},
		{Date: days[1], UserID: users[0], SearchCount: 25, CodeIntelCount: 25},
		{Date: days[1], UserID: users[1], SearchCount: 25, CodeIntelCount: 25},
		{Date: days[0], UserID: users[0], SearchCount: 25, CodeIntelCount: 25},
		{Date: days[0], UserID: users[1], SearchCount: 25, CodeIntelCount: 25},
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Error(diff)
	}
}

func TestEventLogs_SiteUsage(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()

	// This unix timestamp is equivalent to `Friday, May 15, 2020 10:30:00 PM GMT` and is set to
	// be a consistent value so that the tests don't fail when someone runs it at some particular
	// time that falls too near the edge of a week.
	now := time.Unix(1589581800, 0).UTC()

	days := map[time.Time]struct {
		users   []uint32
		names   []string
		sources []string
	}{
		// Today
		now: {
			[]uint32{1, 2, 3, 4, 5},
			[]string{"ViewSiteAdminX"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		// This week
		now.Add(-time.Hour * 24 * 3): {
			[]uint32{0, 2, 3, 5},
			[]string{"ViewRepository", "ViewTree"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		// This week
		now.Add(-time.Hour * 24 * 4): {
			[]uint32{1, 3, 5, 7},
			[]string{"ViewSiteAdminX", "SavedSearchSlackClicked"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		// This month
		now.Add(-time.Hour * 24 * 6): {
			[]uint32{0, 1, 8, 9},
			[]string{"ViewSiteAdminX"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		// This month
		now.Add(-time.Hour * 24 * 12): {
			[]uint32{1, 2, 3, 4, 5, 6, 11},
			[]string{"ViewTree", "SavedSearchSlackClicked"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
		// Previous month
		now.Add(-time.Hour * 24 * 40): {
			[]uint32{0, 1, 5, 6, 13},
			[]string{"SearchResultsQueried", "DiffSearchResultsQueried"},
			[]string{"test", "CODEHOSTINTEGRATION"},
		},
	}

	for day, data := range days {
		for _, user := range data.users {
			for _, name := range data.names {
				for _, source := range data.sources {
					for i := 0; i < 5; i++ {
						e := &Event{
							UserID: user,
							Name:   name,
							URL:    "test",
							Source: source,
							// Jitter current time +/- 30 minutes
							Timestamp: day.Add(time.Minute * time.Duration(rand.Intn(60)-30)),
						}

						if user == 0 {
							e.AnonymousUserID = "deadbeef"
						}

						if err := EventLogs(db).Insert(ctx, e); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
		}
	}

	summary, err := EventLogs(db).siteUsage(ctx, now)
	if err != nil {
		t.Fatal(err)
	}

	expectedSummary := types.SiteUsageSummary{
		Month:                   time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		Week:                    now.Truncate(time.Hour * 24).Add(-time.Hour * 24 * 5), // the previous Sunday
		Day:                     now.Truncate(time.Hour * 24),
		UniquesMonth:            11,
		UniquesWeek:             7,
		UniquesDay:              5,
		RegisteredUniquesMonth:  10,
		RegisteredUniquesWeek:   6,
		RegisteredUniquesDay:    5,
		IntegrationUniquesMonth: 11,
		IntegrationUniquesWeek:  7,
		IntegrationUniquesDay:   5,
		ManageUniquesMonth:      9,
		CodeUniquesMonth:        8,
		VerifyUniquesMonth:      8,
		MonitorUniquesMonth:     0,
		ManageUniquesWeek:       6,
		CodeUniquesWeek:         4,
		VerifyUniquesWeek:       4,
		MonitorUniquesWeek:      0,
	}
	if diff := cmp.Diff(expectedSummary, summary); diff != "" {
		t.Fatal(diff)
	}
}

func TestEventLogs_codeIntelligenceWeeklyUsersCount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()

	names := []string{"codeintel.lsifHover", "codeintel.searchReferences", "unknown event"}
	users1 := []uint32{10, 20, 30, 40, 50, 60, 70, 80}
	users2 := []uint32{15, 25, 35, 45, 55, 65, 75, 85}

	// This unix timestamp is equivalent to `Friday, May 15, 2020 10:30:00 PM GMT` and is set to
	// time that falls too near the edge of a week.
	now := time.Unix(1589581800, 0).UTC()

	for _, name := range names {
		for _, user := range users1 {
			e := &Event{
				UserID: user,
				Name:   name,
				URL:    "test",
				Source: "test",
				// This week; jitter current time +/- 30 minutes
				Timestamp: now.Add(-time.Hour * 24 * 3).Add(time.Minute * time.Duration(rand.Intn(60)-30)),
			}

			if err := EventLogs(db).Insert(ctx, e); err != nil {
				t.Fatal(err)
			}
		}
		for _, user := range users2 {
			e := &Event{
				UserID: user,
				Name:   name,
				URL:    "test",
				Source: "test",
				// This month: jitter current time +/- 30 minutes
				Timestamp: now.Add(-time.Hour * 24 * 12).Add(time.Minute * time.Duration(rand.Intn(60)-30)),
			}

			if err := EventLogs(db).Insert(ctx, e); err != nil {
				t.Fatal(err)
			}
		}
	}

	eventNames := []string{
		"codeintel.lsifHover",
		"codeintel.searchReferences",
	}

	count, err := EventLogs(db).codeIntelligenceWeeklyUsersCount(ctx, eventNames, now)
	if err != nil {
		t.Fatal(err)
	}

	if count != len(users1) {
		t.Errorf("unexpected count. want=%d have=%d", len(users1), count)
	}
}

func TestEventLogs_TestCodeIntelligenceRepositoryCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()
	now := time.Now()

	repos := []struct {
		id        int
		name      string
		deletedAt *time.Time
	}{
		{1, "test01", nil},
		{2, "test02", nil},
		{3, "test03", nil},
		{4, "test04", nil},  // (no LSIF data)
		{5, "test05", &now}, // deleted
	}
	for _, repo := range repos {
		query := sqlf.Sprintf(
			"INSERT INTO repo (id, name, deleted_at) VALUES (%s, %s, %s)",
			repo.id,
			repo.name,
			repo.deletedAt,
		)
		if _, err := db.Exec(query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error preparing database: %s", err.Error())
		}
	}

	uploads := []struct {
		repositoryID int
	}{
		{1},
		{1}, // duplicate
		{2},
		{3},
		{5}, // deleted repository
		{6}, // missing repository
	}
	for i, upload := range uploads {
		query := sqlf.Sprintf(
			"INSERT INTO lsif_uploads (repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (%s, %s, 'idx', 1, '{}', 'completed')",
			upload.repositoryID,
			fmt.Sprintf("%040d", i),
		)
		if _, err := db.Exec(query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error preparing database: %s", err.Error())
		}
	}

	withUploads, withoutUploads, err := EventLogs(db).CodeIntelligenceRepositoryCounts(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if withUploads != 3 {
		t.Errorf("unexpected number of repositories with uploads. want=%d have=%d", 3, withUploads)
	}
	if withoutUploads != 1 {
		t.Errorf("unexpected number of repositories without uploads. want=%d have=%d", 1, withoutUploads)
	}
}

func TestEventLogs_AggregatedCodeIntelEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()

	names := []string{"codeintel.lsifHover", "codeintel.searchReferences.xrepo", "unknown event"}
	users := []uint32{1, 2}

	// This unix timestamp is equivalent to `Friday, May 15, 2020 10:30:00 PM GMT` and is set to
	// be a consistent value so that the tests don't fail when someone runs it at some particular
	// time that falls too near the edge of a week.
	now := time.Unix(1589581800, 0).UTC()

	days := []time.Time{
		now,                           // Today
		now.Add(-time.Hour * 24 * 3),  // This week
		now.Add(-time.Hour * 24 * 4),  // This week
		now.Add(-time.Hour * 24 * 6),  // This month
		now.Add(-time.Hour * 24 * 12), // This month
		now.Add(-time.Hour * 24 * 40), // Previous month
	}

	g, gctx := errgroup.WithContext(ctx)

	for _, user := range users {
		for _, name := range names {
			for _, day := range days {
				for i := 0; i < 25; i++ {
					e := &Event{
						UserID:   user,
						Name:     name,
						URL:      "test",
						Source:   "test",
						Argument: json.RawMessage(fmt.Sprintf(`{"languageId": "lang-%02d"}`, (i%3)+1)),
						// Jitter current time +/- 30 minutes
						Timestamp: day.Add(time.Minute * time.Duration(rand.Intn(60)-30)),
					}

					g.Go(func() error {
						return EventLogs(db).Insert(gctx, e)
					})
				}
			}
		}
	}

	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}

	events, err := EventLogs(db).aggregatedCodeIntelEvents(ctx, now)
	if err != nil {
		t.Fatal(err)
	}

	lang1 := "lang-01"
	lang2 := "lang-02"
	lang3 := "lang-03"

	// the previous Sunday
	week := now.Truncate(time.Hour * 24).Add(-time.Hour * 24 * 5)

	expectedEvents := []types.CodeIntelAggregatedEvent{
		{Name: "codeintel.lsifHover", LanguageID: &lang1, Week: week, TotalWeek: 54, UniquesWeek: 2},
		{Name: "codeintel.lsifHover", LanguageID: &lang2, Week: week, TotalWeek: 48, UniquesWeek: 2},
		{Name: "codeintel.lsifHover", LanguageID: &lang3, Week: week, TotalWeek: 48, UniquesWeek: 2},
		{Name: "codeintel.searchReferences.xrepo", LanguageID: &lang1, Week: week, TotalWeek: 54, UniquesWeek: 2},
		{Name: "codeintel.searchReferences.xrepo", LanguageID: &lang2, Week: week, TotalWeek: 48, UniquesWeek: 2},
		{Name: "codeintel.searchReferences.xrepo", LanguageID: &lang3, Week: week, TotalWeek: 48, UniquesWeek: 2},
	}
	if diff := cmp.Diff(expectedEvents, events); diff != "" {
		t.Fatal(diff)
	}
}

func TestEventLogs_AggregatedSparseCodeIntelEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()

	// This unix timestamp is equivalent to `Friday, May 15, 2020 10:30:00 PM GMT` and is set to
	// be a consistent value so that the tests don't fail when someone runs it at some particular
	// time that falls too near the edge of a week.
	now := time.Unix(1589581800, 0).UTC()

	for i := 0; i < 5; i++ {
		e := &Event{
			UserID:    1,
			Name:      "codeintel.searchReferences.xrepo",
			URL:       "test",
			Source:    "test",
			Argument:  json.RawMessage(fmt.Sprintf(`{"languageId": "lang-%02d"}`, (i%3)+1)),
			Timestamp: now.Add(-time.Hour * 24 * 3), // This week
		}

		if err := EventLogs(db).Insert(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	events, err := EventLogs(db).aggregatedCodeIntelEvents(ctx, now)
	if err != nil {
		t.Fatal(err)
	}

	lang1 := "lang-01"
	lang2 := "lang-02"
	lang3 := "lang-03"

	// the previous Sunday
	week := now.Truncate(time.Hour * 24).Add(-time.Hour * 24 * 5)

	expectedEvents := []types.CodeIntelAggregatedEvent{
		{Name: "codeintel.searchReferences.xrepo", LanguageID: &lang1, Week: week, TotalWeek: 2, UniquesWeek: 1},
		{Name: "codeintel.searchReferences.xrepo", LanguageID: &lang2, Week: week, TotalWeek: 2, UniquesWeek: 1},
		{Name: "codeintel.searchReferences.xrepo", LanguageID: &lang3, Week: week, TotalWeek: 1, UniquesWeek: 1},
	}
	if diff := cmp.Diff(expectedEvents, events); diff != "" {
		t.Fatal(diff)
	}
}

func TestEventLogs_AggregatedSparseSearchEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()

	// This unix timestamp is equivalent to `Friday, May 15, 2020 10:30:00 PM GMT` and is set to
	// be a consistent value so that the tests don't fail when someone runs it at some particular
	// time that falls too near the edge of a week.
	now := time.Unix(1589581800, 0).UTC()

	for i := 0; i < 5; i++ {
		e := &Event{
			UserID: 1,
			Name:   "search.latencies.structural",
			URL:    "test",
			Source: "test",
			// Make durations non-uniform to test percent_cont. The values
			// in this test were hand-checked before being added to the assertion.
			// Adding additional events or changing parameters will require these
			// values to be checked again.
			Argument:  json.RawMessage(fmt.Sprintf(`{"durationMs": %d}`, 50)),
			Timestamp: now.Add(-time.Hour * 24 * 6), // This month
		}

		if err := EventLogs(db).Insert(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	events, err := EventLogs(db).AggregatedSearchEvents(ctx, now)
	if err != nil {
		t.Fatal(err)
	}

	expectedEvents := []types.SearchAggregatedEvent{
		{
			Name:           "search.latencies.structural",
			Month:          time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:           now.Truncate(time.Hour * 24).Add(-time.Hour * 24 * 5), // the previous Sunday
			Day:            now.Truncate(time.Hour * 24),
			TotalMonth:     5,
			TotalWeek:      0,
			TotalDay:       0,
			UniquesMonth:   1,
			UniquesWeek:    0,
			UniquesDay:     0,
			LatenciesMonth: []float64{50, 50, 50},
			LatenciesWeek:  nil,
			LatenciesDay:   nil,
		},
	}
	if diff := cmp.Diff(expectedEvents, events); diff != "" {
		t.Fatal(diff)
	}
}

func TestEventLogs_AggregatedSearchEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()

	names := []string{"search.latencies.literal", "search.latencies.structural", "unknown event"}
	users := []uint32{1, 2}
	durations := []int{40, 65, 72}

	// This unix timestamp is equivalent to `Friday, May 15, 2020 10:30:00 PM GMT` and is set to
	// be a consistent value so that the tests don't fail when someone runs it at some particular
	// time that falls too near the edge of a week.
	now := time.Unix(1589581800, 0).UTC()

	days := []time.Time{
		now,                           // Today
		now.Add(-time.Hour * 24 * 3),  // This week
		now.Add(-time.Hour * 24 * 4),  // This week
		now.Add(-time.Hour * 24 * 6),  // This month
		now.Add(-time.Hour * 24 * 12), // This month
		now.Add(-time.Hour * 24 * 40), // Previous month
	}

	g, gctx := errgroup.WithContext(ctx)

	// add some latencies
	durationOffset := 0
	for _, user := range users {
		for _, name := range names {
			for _, duration := range durations {
				for _, day := range days {
					for i := 0; i < 25; i++ {
						durationOffset++

						e := &Event{
							UserID: user,
							Name:   name,
							URL:    "test",
							Source: "test",
							// Make durations non-uniform to test percent_cont. The values
							// in this test were hand-checked before being added to the assertion.
							// Adding additional events or changing parameters will require these
							// values to be checked again.
							Argument: json.RawMessage(fmt.Sprintf(`{"durationMs": %d}`, duration+durationOffset)),
							// Jitter current time +/- 30 minutes
							Timestamp: day.Add(time.Minute * time.Duration(rand.Intn(60)-30)),
						}

						g.Go(func() error {
							return EventLogs(db).Insert(gctx, e)
						})
					}
				}
			}
		}
	}

	e := &Event{
		UserID: 3,
		Name:   "SearchResultsQueried",
		URL:    "test",
		Source: "test",
		Argument: json.RawMessage(`
{
   "code_search":{
      "query_data":{
         "query":{
             "count_and":3,
             "count_repo_contains_commit_after":2
         },
         "empty":false,
         "combined":"don't care"
      }
   }
}`),
		// Jitter current time +/- 30 minutes
		Timestamp: now.Add(-time.Hour * 24 * 3).Add(time.Minute * time.Duration(rand.Intn(60)-30)),
	}

	if err := EventLogs(db).Insert(gctx, e); err != nil {
		t.Fatal(err)
	}

	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}

	events, err := EventLogs(db).AggregatedSearchEvents(ctx, now)
	if err != nil {
		t.Fatal(err)
	}

	expectedEvents := []types.SearchAggregatedEvent{
		{
			Name:           "search.latencies.literal",
			Month:          time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:           now.Truncate(time.Hour * 24).Add(-time.Hour * 24 * 5), // the previous Sunday
			Day:            now.Truncate(time.Hour * 24),
			TotalMonth:     int32(len(users) * len(durations) * 25 * 5), // 5 days in month
			TotalWeek:      int32(len(users) * len(durations) * 25 * 3), // 3 days in week
			TotalDay:       int32(len(users) * len(durations) * 25),
			UniquesMonth:   2,
			UniquesWeek:    2,
			UniquesDay:     2,
			LatenciesMonth: []float64{944, 1772.1, 1839.51},
			LatenciesWeek:  []float64{919, 1752.1, 1792.51},
			LatenciesDay:   []float64{894, 1732.1, 1745.51},
		},
		{
			Name:           "search.latencies.structural",
			Month:          time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:           now.Truncate(time.Hour * 24).Add(-time.Hour * 24 * 5), // the previous Sunday
			Day:            now.Truncate(time.Hour * 24),
			TotalMonth:     int32(len(users) * len(durations) * 25 * 5), // 5 days in month
			TotalWeek:      int32(len(users) * len(durations) * 25 * 3), // 3 days in week
			TotalDay:       int32(len(users) * len(durations) * 25),
			UniquesMonth:   2,
			UniquesWeek:    2,
			UniquesDay:     2,
			LatenciesMonth: []float64{1394, 2222.1, 2289.51},
			LatenciesWeek:  []float64{1369, 2202.1, 2242.51},
			LatenciesDay:   []float64{1344, 2182.1, 2195.51},
		},
		{
			Name:         "count_and",
			Month:        time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:         now.Truncate(time.Hour * 24).Add(-time.Hour * 24 * 5),
			Day:          now.Truncate(time.Hour * 24),
			TotalMonth:   3,
			TotalWeek:    3,
			TotalDay:     0,
			UniquesMonth: 1,
			UniquesWeek:  1,
		},
		{
			Name:         "count_repo_contains_commit_after",
			Month:        time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
			Week:         now.Truncate(time.Hour * 24).Add(-time.Hour * 24 * 5),
			Day:          now.Truncate(time.Hour * 24),
			TotalMonth:   2,
			TotalWeek:    2,
			TotalDay:     0,
			UniquesMonth: 1,
			UniquesWeek:  1,
		},
	}
	if diff := cmp.Diff(expectedEvents, events); diff != "" {
		t.Fatal(diff)
	}
}

func TestEventLogs_ListAll(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()

	now := time.Now()

	startDate, _ := calcStartDate(now, Daily, 3)

	events := []*Event{
		{
			UserID:    1,
			Name:      "SearchResultsQueried",
			URL:       "test",
			Source:    "test",
			Timestamp: startDate,
		}, {
			UserID:    2,
			Name:      "codeintel",
			URL:       "test",
			Source:    "test",
			Timestamp: startDate,
		},
		{
			UserID:    2,
			Name:      "ViewRepository",
			URL:       "test",
			Source:    "test",
			Timestamp: startDate,
		},
		{
			UserID:    2,
			Name:      "SearchResultsQueried",
			URL:       "test",
			Source:    "test",
			Timestamp: startDate,
		}}

	for _, event := range events {
		if err := EventLogs(db).Insert(ctx, event); err != nil {
			t.Fatal(err)
		}
	}

	searchResultQueriedEvent := "SearchResultsQueried"
	have, err := EventLogs(db).ListAll(ctx, EventLogsListOptions{EventName: &searchResultQueriedEvent})
	if err != nil {
		t.Fatal(err)
	}

	want := 2

	if diff := cmp.Diff(want, len(have)); diff != "" {
		t.Error(diff)
	}
}

func TestEventLogs_LatestPing(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")

	t.Run("with no pings in database", func(t *testing.T) {
		ctx := context.Background()
		ping, err := EventLogs(db).LatestPing(ctx)
		if ping != nil {
			t.Fatalf("have ping %+v, expected nil", ping)
		}
		if err != sql.ErrNoRows {
			t.Fatalf("have err %+v, expected no rows error", err)
		}
	})

	t.Run("with existing pings in database", func(t *testing.T) {
		userID := int32(0)
		timestamp := timeutil.Now()

		ctx := context.Background()
		events := []*Event{
			{
				UserID:          0,
				Name:            "ping",
				URL:             "test",
				AnonymousUserID: "test",
				Source:          "test",
				Timestamp:       timestamp,
				Argument:        json.RawMessage(`{"key": "value1"}`),
			}, {
				UserID:          0,
				Name:            "ping",
				URL:             "test",
				AnonymousUserID: "test",
				Source:          "test",
				Timestamp:       timestamp,
				Argument:        json.RawMessage(`{"key": "value2"}`),
			},
		}
		for _, event := range events {
			if err := EventLogs(db).Insert(ctx, event); err != nil {
				t.Fatal(err)
			}
		}

		gotPing, err := EventLogs(db).LatestPing(ctx)
		if err != nil || gotPing == nil {
			t.Fatal(err)
		}
		expectedPing := &types.Event{
			ID:              2,
			Name:            events[1].Name,
			URL:             events[1].URL,
			UserID:          &userID,
			AnonymousUserID: events[1].AnonymousUserID,
			Version:         version.Version(),
			Argument:        string(events[1].Argument),
			Source:          events[1].Source,
			Timestamp:       timestamp,
		}
		if diff := cmp.Diff(gotPing, expectedPing); diff != "" {
			t.Fatal(diff)
		}
	})
}

// makeTestEvent sets the required (uninteresting) fields that are required on insertion
// due to database constraints. This method will also add some sub-day jitter to the timestamp.
func makeTestEvent(e *Event) *Event {
	if e.UserID == 0 {
		e.UserID = 1
	}
	e.Name = "foo"
	e.URL = "test"
	e.Source = "WEB"
	e.Timestamp = e.Timestamp.Add(time.Minute * time.Duration(rand.Intn(60*12)))
	return e
}

func assertUsageValue(t *testing.T, v UsageValue, start time.Time, count int) {
	if v.Start != start {
		t.Errorf("got Start %q, want %q", v.Start, start)
	}
	if v.Count != count {
		t.Errorf("got Count %d, want %d", v.Count, count)
	}
}
