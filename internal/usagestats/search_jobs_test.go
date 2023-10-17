package usagestats

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSearchJobsUsageStatistics(t *testing.T) {
	ctx := context.Background()

	defer func() {
		timeNow = time.Now
	}()

	now := time.Date(2021, 1, 28, 0, 0, 0, 0, time.UTC)
	mockTimeNow(now)

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO event_logs
			(id, name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
		VALUES
			(1, 'ViewSearchJobsListPage', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(2, 'ViewSearchJobsListPage', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(3, 'SearchJobsValidationErrors', '{"errors": ["rev", "or_operator"]}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(4, 'SearchJobsValidationErrors', '{"errors": ["rev", "or_operator"]}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(5, 'SearchJobsValidationErrors', '{"errors": ["and", "or_operator"]}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(6, 'SearchJobsValidationErrors', '{"errors": ["and"]}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(7, 'SearchJobsResultDownloadClick', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc19ac', 'WEB', '3.23.0', $1::timestamp - interval '2 days'),
			(8, 'SearchJobsResultViewLogsClick', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '2 days'),
			(9, 'SearchJobsCreateClick', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(10, 'SearchJobsSearchFormShown', '{"validState": "valid"}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day'),
			(11, 'SearchJobsSearchFormShown', '{"validState": "invalid"}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestamp - interval '1 day')
	`, now)

	if err != nil {
		t.Fatal(err)
	}

	have, err := GetSearchJobsUsageStatistics(ctx, db)

	if err != nil {
		t.Fatal(err)
	}

	oneInt := int32(1)
	twoInt := int32(2)

	weeklySearchJobsSearchFormShown := []types.SearchJobsSearchFormShownPing{
		{
			ValidState: "invalid",
			TotalCount: 1,
		},
		{
			ValidState: "valid",
			TotalCount: 1,
		},
	}

	weeklySearchJobsValidationErrors := []types.SearchJobsValidationErrorPing{
		{
			TotalCount: 2,
			Errors:     []string{"rev", "or_operator"},
		},
		{
			TotalCount: 1,
			Errors:     []string{"and", "or_operator"},
		},
		{
			TotalCount: 1,
			Errors:     []string{"and"},
		},
	}

	want := &types.SearchJobsUsageStatistics{
		WeeklySearchJobsPageViews:            &twoInt,
		WeeklySearchJobsCreateClick:          &oneInt,
		WeeklySearchJobsDownloadClicks:       &oneInt,
		WeeklySearchJobsViewLogsClicks:       &oneInt,
		WeeklySearchJobsUniquePageViews:      &oneInt,
		WeeklySearchJobsUniqueDownloadClicks: &oneInt,
		WeeklySearchJobsUniqueViewLogsClicks: &oneInt,
		WeeklySearchJobsSearchFormShown:      []types.SearchJobsSearchFormShownPing{},
		WeeklySearchJobsValidationErrors:     []types.SearchJobsValidationErrorPing{},
	}

	want.WeeklySearchJobsSearchFormShown = weeklySearchJobsSearchFormShown
	want.WeeklySearchJobsValidationErrors = weeklySearchJobsValidationErrors

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}
