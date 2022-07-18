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

func TestGetAdminAnalyticsUsageStatistics(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	now := time.Now()

	_, err := db.ExecContext(context.Background(), `
INSERT INTO event_logs
	(id, name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
VALUES
	(1, 'AdminAnalyticsSearchViewed', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(2, 'AdminAnalyticsCodeIntelViewed', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(3, 'AdminAnalyticsUsersViewed', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(4, 'AdminAnalyticsBatchChangesViewed', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(5, 'AdminAnalyticsNotebooksViewed', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(6, 'AdminAnalyticsSearchPercentageInputEdited', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(7, 'AdminAnalyticsSearchMinutesInputEdited', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(8, 'AdminAnalyticsCodeIntelPercentageInputEdited', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(9, 'AdminAnalyticsCodeIntelMinutesInputEdited', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(10, 'AdminAnalyticsBatchChangesPercentageInputEdited', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(11, 'AdminAnalyticsBatchChangesMinutesInputEdited', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(12, 'AdminAnalyticsNotebooksPercentageInputEdited', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(13, 'AdminAnalyticsNotebooksMinutesInputEdited', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(14, 'AdminAnalyticsSearchDateRangeLAST_WEEKSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(15, 'AdminAnalyticsSearchDateRangeLAST_MONTHSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(16, 'AdminAnalyticsSearchDateRangeLAST_THREE_MONTHSSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(17, 'AdminAnalyticsCodeIntelDateRangeLAST_WEEKSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(18, 'AdminAnalyticsCodeIntelDateRangeLAST_MONTHSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(19, 'AdminAnalyticsCodeIntelDateRangeLAST_THREE_MONTHSSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(20, 'AdminAnalyticsUsersDateRangeLAST_WEEKSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(21, 'AdminAnalyticsUsersDateRangeLAST_MONTHSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(22, 'AdminAnalyticsUsersDateRangeLAST_THREE_MONTHSSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(23, 'AdminAnalyticsBatchChangesDateRangeLAST_WEEKSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(24, 'AdminAnalyticsBatchChangesDateRangeLAST_MONTHSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(25, 'AdminAnalyticsBatchChangesDateRangeLAST_THREE_MONTHSSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(26, 'AdminAnalyticsNotebooksDateRangeLAST_WEEKSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(27, 'AdminAnalyticsNotebooksDateRangeLAST_MONTHSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(28, 'AdminAnalyticsNotebooksDateRangeLAST_THREE_MONTHSSelected', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(29, 'AdminAnalyticsSearchAggTotalsClicked', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(30, 'AdminAnalyticsSearchAggUniquesClicked', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(31, 'AdminAnalyticsCodeIntelAggTotalsClicked', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(32, 'AdminAnalyticsCodeIntelAggUniquesClicked', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(33, 'AdminAnalyticsUsersAggTotalsClicked', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(34, 'AdminAnalyticsUsersAggUniquesClicked', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(35, 'AdminAnalyticsNotebooksAggTotalsClicked', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(36, 'AdminAnalyticsNotebooksAggUniquesClicked', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day')
`, now)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetAdminAnalyticsUsageStatistics(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	count := int32(1)

	want := &types.AdminAnalyticsUsageStatistics{
		AdminAnalyticsSearchPageViews:                                &count,
		AdminAnalyticsCodeIntelPageViews:                             &count,
		AdminAnalyticsUsersPageViews:                                 &count,
		AdminAnalyticsBatchChangesPageViews:                          &count,
		AdminAnalyticsNotebooksPageViews:                             &count,
		AdminAnalyticsSearchPercentageInputEdited:                    &count,
		AdminAnalyticsSearchMinutesInputEdited:                       &count,
		AdminAnalyticsCodeIntelPercentageInputEdited:                 &count,
		AdminAnalyticsCodeIntelMinutesInputEdited:                    &count,
		AdminAnalyticsBatchChangesPercentageInputEdited:              &count,
		AdminAnalyticsBatchChangesMinutesInputEdited:                 &count,
		AdminAnalyticsNotebooksPercentageInputEdited:                 &count,
		AdminAnalyticsNotebooksMinutesInputEdited:                    &count,
		AdminAnalyticsSearchDateRangeLAST_WEEKSelected:               &count,
		AdminAnalyticsSearchDateRangeLAST_MONTHSelected:              &count,
		AdminAnalyticsSearchDateRangeLAST_THREE_MONTHSSelected:       &count,
		AdminAnalyticsCodeIntelDateRangeLAST_WEEKSelected:            &count,
		AdminAnalyticsCodeIntelDateRangeLAST_MONTHSelected:           &count,
		AdminAnalyticsCodeIntelDateRangeLAST_THREE_MONTHSSelected:    &count,
		AdminAnalyticsUsersDateRangeLAST_WEEKSelected:                &count,
		AdminAnalyticsUsersDateRangeLAST_MONTHSelected:               &count,
		AdminAnalyticsUsersDateRangeLAST_THREE_MONTHSSelected:        &count,
		AdminAnalyticsBatchChangesDateRangeLAST_WEEKSelected:         &count,
		AdminAnalyticsBatchChangesDateRangeLAST_MONTHSelected:        &count,
		AdminAnalyticsBatchChangesDateRangeLAST_THREE_MONTHSSelected: &count,
		AdminAnalyticsNotebooksDateRangeLAST_WEEKSelected:            &count,
		AdminAnalyticsNotebooksDateRangeLAST_MONTHSelected:           &count,
		AdminAnalyticsNotebooksDateRangeLAST_THREE_MONTHSSelected:    &count,
		AdminAnalyticsSearchAggTotalsClicked:                         &count,
		AdminAnalyticsSearchAggUniquesClicked:                        &count,
		AdminAnalyticsCodeIntelAggTotalsClicked:                      &count,
		AdminAnalyticsCodeIntelAggUniquesClicked:                     &count,
		AdminAnalyticsUsersAggTotalsClicked:                          &count,
		AdminAnalyticsUsersAggUniquesClicked:                         &count,
		AdminAnalyticsNotebooksAggTotalsClicked:                      &count,
		AdminAnalyticsNotebooksAggUniquesClicked:                     &count,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}
