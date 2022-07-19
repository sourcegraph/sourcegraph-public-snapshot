package adminanalytics

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestNotebooksCreationsLastWeek(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	now := bod(time.Now())

	_, err := db.ExecContext(context.Background(), `
INSERT INTO event_logs
	(id, name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
VALUES
	(1, 'SearchNotebookCreated', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ea', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(2, 'SearchNotebookCreated', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc91eb', 'WEB', 'version', $1::timestamp - interval '2 day'),
	(3, 'SearchNotebookCreated', '{}', '', 0, '420657f0-d443-4d16-ac7d-003d8cdc91ec', 'WEB', 'version', $1::timestamp - interval '2 day'),
	(4, 'SearchNotebookCreated', '{}', '', 0, '420657f0-d443-4d16-ac7d-003d8cdc91ec', 'WEB', 'version', $1::timestamp - interval '20 day'),
	(5, 'SearchNotebookCreated', '{}', '', 0, '420657f0-d443-4d16-ac7d-003d8cdc91ec', 'WEB', 'version', $1::timestamp + interval '1 day')
	`, now)
	if err != nil {
		t.Fatal(err)
	}

	noSetCache := true
	store := Notebooks{
		DateRange:  "LAST_WEEK",
		DB:         db,
		Cache:      false,
		NoSetCache: &noSetCache,
	}

	fetcher, err := store.Creations()
	if err != nil {
		t.Fatal(err)
	}

	results, err := fetcher.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 8 {
		t.Fatal(fmt.Sprintf("%d nodes returned rather than 8", len(results)))
	}

	nodes := []*AnalyticsNode{
		{
			Data: AnalyticsNodeData{
				Date:            now.AddDate(0, 0, -1),
				Count:           1,
				UniqueUsers:     1,
				RegisteredUsers: 1,
			},
		},
		{
			Data: AnalyticsNodeData{
				Date:            now.AddDate(0, 0, -2),
				Count:           2,
				UniqueUsers:     2,
				RegisteredUsers: 1,
			},
		},
	}

	for _, node := range nodes {
		var found *AnalyticsNode

		for _, result := range results {
			if bod(node.Data.Date).Equal(bod(result.Data.Date)) {
				found = result
			}
		}

		if diff := cmp.Diff(node, found); diff != "" {
			t.Fatal(diff)
		}
	}

	summaryResult, err := fetcher.Summary(ctx)
	if err != nil {
		t.Fatal(err)
	}

	summary := &AnalyticsSummary{
		Data: AnalyticsSummaryData{
			TotalCount:           3,
			TotalUniqueUsers:     3,
			TotalRegisteredUsers: 2,
		},
	}

	if diff := cmp.Diff(summary, summaryResult); diff != "" {
		t.Fatal(diff)
	}
}

func TestNotebooksCreationsLastMonth(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	now := bod(time.Now())

	_, err := db.ExecContext(context.Background(), `
INSERT INTO event_logs
	(id, name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
VALUES
	(1, 'SearchNotebookCreated', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ea', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(2, 'SearchNotebookCreated', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc91eb', 'WEB', 'version', $1::timestamp - interval '10 day'),
	(3, 'SearchNotebookCreated', '{}', '', 0, '420657f0-d443-4d16-ac7d-003d8cdc91ec', 'WEB', 'version', $1::timestamp - interval '10 day'),
	(4, 'SearchNotebookCreated', '{}', '', 0, '420657f0-d443-4d16-ac7d-003d8cdc91ec', 'WEB', 'version', $1::timestamp - interval '100 day'),
	(5, 'SearchNotebookCreated', '{}', '', 0, '420657f0-d443-4d16-ac7d-003d8cdc91ec', 'WEB', 'version', $1::timestamp + interval '1 day')
	`, now)
	if err != nil {
		t.Fatal(err)
	}

	noSetCache := true
	store := Notebooks{
		DateRange:  "LAST_MONTH",
		DB:         db,
		Cache:      false,
		NoSetCache: &noSetCache,
	}

	fetcher, err := store.Creations()
	if err != nil {
		t.Fatal(err)
	}

	results, err := fetcher.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) < 28 {
		t.Fatal(fmt.Sprintf("only %d nodes returned", len(results)))
	}

	nodes := []*AnalyticsNode{
		{
			Data: AnalyticsNodeData{
				Date:            now.AddDate(0, 0, -1),
				Count:           1,
				UniqueUsers:     1,
				RegisteredUsers: 1,
			},
		},
		{
			Data: AnalyticsNodeData{
				Date:            now.AddDate(0, 0, -10),
				Count:           2,
				UniqueUsers:     2,
				RegisteredUsers: 1,
			},
		},
	}

	for _, node := range nodes {
		var found *AnalyticsNode

		for _, result := range results {
			if bod(node.Data.Date).Equal(bod(result.Data.Date)) {
				found = result
			}
		}

		if diff := cmp.Diff(node, found); diff != "" {
			t.Fatal(diff)
		}
	}

	summaryResult, err := fetcher.Summary(ctx)
	if err != nil {
		t.Fatal(err)
	}

	summary := &AnalyticsSummary{
		Data: AnalyticsSummaryData{
			TotalCount:           3,
			TotalUniqueUsers:     3,
			TotalRegisteredUsers: 2,
		},
	}

	if diff := cmp.Diff(summary, summaryResult); diff != "" {
		t.Fatal(diff)
	}
}

func TestNotebooksCreationsLastThreeMonths(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	now := bod(time.Now())

	_, err := db.ExecContext(context.Background(), `
INSERT INTO event_logs
	(id, name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
VALUES
	(1, 'SearchNotebookCreated', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ea', 'WEB', 'version', $1::timestamp),
	(2, 'SearchNotebookCreated', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc91eb', 'WEB', 'version', $1::timestamp),
	(3, 'SearchNotebookCreated', '{}', '', 0, '420657f0-d443-4d16-ac7d-003d8cdc91ec', 'WEB', 'version', $1::timestamp),
	(4, 'SearchNotebookCreated', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ea', 'WEB', 'version', $1::timestamp - interval '40 day'),
	(5, 'SearchNotebookCreated', '{}', '', 2, '420657f0-d443-4d16-ac7d-003d8cdc91eb', 'WEB', 'version', $1::timestamp - interval '40 day'),
	(6, 'SearchNotebookCreated', '{}', '', 0, '420657f0-d443-4d16-ac7d-003d8cdc91ec', 'WEB', 'version', $1::timestamp - interval '40 day'),
	(7, 'SearchNotebookCreated', '{}', '', 0, '420657f0-d443-4d16-ac7d-003d8cdc91ec', 'WEB', 'version', $1::timestamp - interval '100 day'),
	(8, 'SearchNotebookCreated', '{}', '', 0, '420657f0-d443-4d16-ac7d-003d8cdc91ec', 'WEB', 'version', $1::timestamp + interval '1 day')
	`, now)
	if err != nil {
		t.Fatal(err)
	}

	noSetCache := true
	store := Notebooks{
		DateRange:  "LAST_THREE_MONTHS",
		DB:         db,
		Cache:      false,
		NoSetCache: &noSetCache,
	}

	fetcher, err := store.Creations()
	if err != nil {
		t.Fatal(err)
	}

	results, err := fetcher.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) < 12 {
		t.Fatal(fmt.Sprintf("only %d nodes returned", len(results)))
	}

	nodes := []*AnalyticsNode{
		{
			Data: AnalyticsNodeData{
				Date:            sow(now),
				Count:           3,
				UniqueUsers:     3,
				RegisteredUsers: 2,
			},
		},
		{
			Data: AnalyticsNodeData{
				Date:            sow(now.AddDate(0, 0, -40)),
				Count:           3,
				UniqueUsers:     3,
				RegisteredUsers: 2,
			},
		},
	}

	for _, node := range nodes {
		var found *AnalyticsNode

		for _, result := range results {
			if bod(node.Data.Date).Equal(bod(result.Data.Date)) {
				found = result
			}
		}

		if diff := cmp.Diff(node, found); diff != "" {
			t.Fatal(diff)
		}
	}

	summaryResult, err := fetcher.Summary(ctx)
	if err != nil {
		t.Fatal(err)
	}

	summary := &AnalyticsSummary{
		Data: AnalyticsSummaryData{
			TotalCount:           6,
			TotalUniqueUsers:     3,
			TotalRegisteredUsers: 2,
		},
	}

	if diff := cmp.Diff(summary, summaryResult); diff != "" {
		t.Fatal(diff)
	}
}
