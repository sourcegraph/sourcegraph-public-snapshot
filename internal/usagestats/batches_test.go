package usagestats

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGetBatchChangesUsageStatistics(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	// Create stub repo.
	repoStore := db.Repos()
	esStore := db.ExternalServices()

	// making use of a mock clock here to ensure all time operations are appropriately mocked
	// https://docs-legacy.sourcegraph.com/dev/background-information/languages/testing_go_code#testing-time
	clock := glock.NewMockClock()
	now := clock.Now()

	svc := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "token": "beef", "repos": ["owner/repo"]}`),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := esStore.Upsert(ctx, &svc); err != nil {
		t.Fatalf("failed to insert external services: %v", err)
	}
	repo := &types.Repo{
		Name: "test/repo",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          fmt.Sprintf("external-id-%d", svc.ID),
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*types.SourceInfo{
			svc.URN(): {
				ID:       svc.URN(),
				CloneURL: "https://secrettoken@test/repo",
			},
		},
	}
	if err := repoStore.Create(ctx, repo); err != nil {
		t.Fatal(err)
	}

	// Create a user.
	user, err := db.Users().Create(ctx, database.NewUser{Username: "test"})
	if err != nil {
		t.Fatal(err)
	}

	// Create another user.
	user2, err := db.Users().Create(ctx, database.NewUser{Username: "test-2"})
	if err != nil {
		t.Fatal(err)
	}

	// Due to irregularity in the amount of days in a month, subtracting simply a month from a date can deduct
	// 30 days, but that's incorrect because not every month has 30 days.
	// This poses a problem, therefore deducting three days after the initial month deduction ensures we'll
	// always get a date that falls in the previous month regardless of the day in question.
	lastMonthCreationDate := now.AddDate(0, -1, -3)
	twoMonthsAgoCreationDate := now.AddDate(0, -2, -3)

	// Create batch specs
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO batch_specs
			(id, rand_id, raw_spec, namespace_user_id, user_id, created_from_raw, created_at)
		VALUES
		    -- 3 from this month, 2 from executors by the same user
			(1, '123', '{}', $1, $1, FALSE, $3::timestamp),
			(2, '157', '{}', $2, $2, TRUE, $3::timestamp),
			(3, 'U93', '{}', $2, $2, TRUE, $3::timestamp),
			-- 3 from last month, 2 from executors by different users
			(4, '456', '{}', $1, $1, FALSE, $4::timestamp),
			(5, '789', '{}', $1, $1, TRUE, $4::timestamp),
			(6, 'C80', '{}', $2, $2, TRUE, $4::timestamp),
			-- 1 from two months ago, from executors
			(7, 'KEK', '{}', $2, $2, TRUE, $5::timestamp)
	`, user.ID, user2.ID, now, lastMonthCreationDate, twoMonthsAgoCreationDate)
	if err != nil {
		t.Fatal(err)
	}

	// Create batch spec workspaces
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO batch_spec_workspaces
			(id, repo_id, batch_spec_id, branch, commit, path, file_matches)
		VALUES
			(1, 1, 2, 'refs/heads/main', 'some-commit', '', '{README.md}'),
			(2, 1, 2, 'refs/heads/main', 'some-commit', '', '{README.md}'),
			(3, 1, 3, 'refs/heads/main', 'some-commit', '', '{README.md}'),
			(4, 1, 5, 'refs/heads/main', 'some-commit', '', '{README.md}'),
			(5, 1, 7, 'refs/heads/main', 'some-commit', '', '{README.md}')
	`)
	if err != nil {
		t.Fatal(err)
	}

	workspaceExecutionStartedDate := now.Add(-10 * time.Minute) // 10 minutes ago

	lastMonthWorkspaceExecutionStartedDate := now.AddDate(0, -1, 2)                                         // Over a month ago
	lastMonthWorkspaceExecutionFinishedDate := lastMonthWorkspaceExecutionStartedDate.Add(10 * time.Minute) // 10 minutes later

	// Create batch spec workspace execution jobs
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO batch_spec_workspace_execution_jobs
			(id, batch_spec_workspace_id, user_id, started_at, finished_at)
		VALUES
			-- Finished this month
			(1, 1, $6, $4::timestamp, $3::timestamp),
			(2, 2, $6, $4::timestamp, $3::timestamp),
			(3, 3, $6, $4::timestamp, $3::timestamp),
			-- Finished last month
			(4, 4, $5, $1::timestamp, $2::timestamp),
			-- Processing: has been started but not finished
			(5, 3, $6, $4::timestamp, NULL),
			-- Queued: has not been started or finished
			(6, 3, $6, NULL, NULL),
			(7, 3, $6, NULL, NULL)
	`, lastMonthWorkspaceExecutionStartedDate, lastMonthWorkspaceExecutionFinishedDate, now, workspaceExecutionStartedDate, user.ID, user2.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Create event logs
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO event_logs
			(id, name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
		VALUES
		-- User 23, created a batch change last month and closes it
			(3, 'BatchSpecCreated', '{"changeset_specs_count": 3}', '', 23, '', 'backend', 'version', date_trunc('month', CURRENT_DATE) - INTERVAL '2 days'),
			(4, 'BatchSpecCreated', '{"changeset_specs_count": 1}', '', 23, '', 'backend', 'version', date_trunc('month', CURRENT_DATE) - INTERVAL '2 days'),
			(5, 'BatchSpecCreated', '{}', '', 23, '', 'backend', 'version', date_trunc('month', CURRENT_DATE) - INTERVAL '2 days'),
			(6, 'ViewBatchChangeApplyPage', '{}', 'https://sourcegraph.test:3443/users/mrnugget/batch-changes/apply/RANDID', 23, '5d302f47-9e91-4b3d-9e96-469b5601a765', 'WEB', 'version', date_trunc('month', CURRENT_DATE) - INTERVAL '2 days'),
			(7, 'BatchChangeCreated', '{"batch_change_id": 1}', '', 23, '', 'backend', 'version', date_trunc('month', CURRENT_DATE) - INTERVAL '2 days'),
			(8, 'ViewBatchChangeDetailsPageAfterCreate', '{}', 'https://sourcegraph.test:3443/users/mrnugget/batch-changes/gitignore-files', 23, '5d302f47-9e91-4b3d-9e96-469b5601a765', 'WEB', 'version', date_trunc('month', CURRENT_DATE) - INTERVAL '2 days'),
			(9, 'ViewBatchChangeDetailsPageAfterUpdate', '{}', 'https://sourcegraph.test:3443/users/mrnugget/batch-changes/gitignore-files', 23, '5d302f47-9e91-4b3d-9e96-469b5601a765', 'WEB', 'version', date_trunc('month', CURRENT_DATE) - INTERVAL '2 days'),
			(10, 'ViewBatchChangeDetailsPage', '{}', 'https://sourcegraph.test:3443/users/mrnugget/batch-changes/gitignore-files', 23, '5d302f47-9e91-4b3d-9e96-469b5601a765', 'WEB', 'version', date_trunc('month', CURRENT_DATE) - INTERVAL '2 days'),
			(11, 'BatchChangeCreatedOrUpdated', '{"batch_change_id": 1}', '', 23, '', 'backend', 'version', date_trunc('month', CURRENT_DATE) - INTERVAL '2 days'),
			(12, 'BatchChangeClosed', '{"batch_change_id": 1}', '', 23, '', 'backend', 'version', date_trunc('month', CURRENT_DATE) - INTERVAL '2 days'),
			(13, 'BatchChangeDeleted', '{"batch_change_id": 1}', '', 23, '', 'backend', 'version', date_trunc('month', CURRENT_DATE) - INTERVAL '2 days'),
		-- User 24, created a batch change today and closes it
			(16, 'BatchSpecCreated', '{}', '', 24, '', 'backend', 'version', $1::timestamp),
			(17, 'ViewBatchChangeApplyPage', '{}', 'https://sourcegraph.test:3443/users/mrnugget/batch-changes/apply/RANDID-2', 24, '5d302f47-9e91-4b3d-9e96-469b5601a765', 'WEB', 'version', $1::timestamp),
			(18, 'BatchChangeCreated', '{"batch_change_id": 2}', '', 24, '', 'backend', 'version', $1::timestamp),
			(19, 'ViewBatchChangeDetailsPageAfterCreate', '{}', 'https://sourcegraph.test:3443/users/mrnugget/batch-changes/foobar-files', 24, '5d302f47-9e91-4b3d-9e96-469b5601a765', 'WEB', 'version', $1::timestamp),
			(20, 'ViewBatchChangeDetailsPageAfterUpdate', '{}', 'https://sourcegraph.test:3443/users/mrnugget/batch-changes/foobar-files', 24, '5d302f47-9e91-4b3d-9e96-469b5601a765', 'WEB', 'version', $1::timestamp),
			(21, 'BatchChangeCreatedOrUpdated', '{"batch_change_id": 2}', '', 24, '', 'backend', 'version', $1::timestamp),
			(22, 'BatchChangeClosed', '{"batch_change_id": 2}', '', 24, '', 'backend', 'version', $1::timestamp),
			(23, 'BatchChangeDeleted', '{"batch_change_id": 2}', '', 24, '', 'backend', 'version', $1::timestamp),
		-- User 25, only views the batch change, today
			(29, 'ViewBatchChangeDetailsPage', '{}', 'https://sourcegraph.test:3443/users/mrnugget/batch-changes/gitignore-files', 25, '5d302f47-9e91-4b3d-9e96-469b5601a765', 'WEB', 'version', $1::timestamp),
			(30, 'ViewBatchChangesListPage', '{}', 'https://sourcegraph.test:3443/users/mrnugget/batch-changes', 25, '5d302f47-9e91-4b3d-9e96-469b5601a765', 'WEB', 'version', $1::timestamp),
			(31, 'ViewBatchChangeDetailsPage', '{}', 'https://sourcegraph.test:3443/users/mrnugget/batch-changes/foobar-files', 25, '5d302f47-9e91-4b3d-9e96-469b5601a765', 'WEB', 'version', $1::timestamp)
	`, now)
	if err != nil {
		t.Fatal(err)
	}

	batchChangeCreationDate1 := now.Add(-24 * 7 * 8 * time.Hour)  // 8 weeks ago
	batchChangeCreationDate2 := now.Add(-24 * 3 * time.Hour)      // 3 days ago
	batchChangeCreationDate3 := now.Add(-24 * 7 * 60 * time.Hour) // 60 weeks ago

	// Create batch changes
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO batch_changes
			(id, name, batch_spec_id, created_at, last_applied_at, namespace_user_id, closed_at)
		VALUES
			(1, 'test',   1, $2::timestamp, $5::timestamp, $1, NULL),
			(2, 'test-2', 4, $3::timestamp, $5::timestamp, $1, $5::timestamp),
			(3, 'test-3', 5, $4::timestamp, $5::timestamp, $1, NULL)
	`, user.ID, batchChangeCreationDate1, batchChangeCreationDate2, batchChangeCreationDate3, now)
	if err != nil {
		t.Fatal(err)
	}

	changesetIDOne := 1
	changesetIDTwo := 2
	changesetIDFour := 4
	changesetIDFive := 5
	changesetIDSix := 6

	// Create 6 changesets.
	// 2 tracked: one OPEN, one MERGED.
	// 4 created by a batch change: 2 open (one with diffstat, one without), 2 merged (one with diffstat, one without)
	// missing diffstat shouldn't happen anymore (due to migration), but it's still a nullable field.
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO changesets
			(id, repo_id, external_service_type, owned_by_batch_change_id, batch_change_ids, external_state, publication_state, diff_stat_added, diff_stat_deleted)
		VALUES
		    -- tracked
			($2, $1, 'github', NULL, '{"1": {"detached": false}}', 'OPEN',   'PUBLISHED', 16, 12),
			($3, $1, 'github', NULL, '{"2": {"detached": false}}', 'MERGED', 'PUBLISHED', 16, 14),
			-- created by batch change
			($4,  $1, 'github', 1, '{"1": {"detached": false}}', 'OPEN',   'PUBLISHED', 12, 16),
			($5,  $1, 'github', 1, '{"1": {"detached": false}}', 'OPEN',   'PUBLISHED', NULL, NULL),
			($6,  $1, 'github', 1, '{"1": {"detached": false}}', 'DRAFT',  'PUBLISHED', NULL, NULL),
			(7,  $1, 'github', 2, '{"2": {"detached": false}}',  NULL,    'UNPUBLISHED', 16, 12),
			(8,  $1, 'github', 2, '{"2": {"detached": false}}', 'MERGED', 'PUBLISHED', 16, 12),
			(9,  $1, 'github', 2, '{"2": {"detached": false}}', 'MERGED', 'PUBLISHED', NULL, NULL),
			(10, $1, 'github', 2, '{"2": {"detached": false}}',  NULL,    'UNPUBLISHED', 16, 12),
			(11, $1, 'github', 2, '{"2": {"detached": false}}', 'CLOSED', 'PUBLISHED', NULL, NULL),
			(12, $1, 'github', 3, '{"3": {"detached": false}}', 'OPEN',   'PUBLISHED', 12, 16),
			(13, $1, 'github', 3, '{"3": {"detached": false}}', 'OPEN',   'PUBLISHED', NULL, NULL)
	`, repo.ID, changesetIDOne, changesetIDTwo, changesetIDFour, changesetIDFive, changesetIDSix)
	if err != nil {
		t.Fatal(err)
	}

	// inactive executors last seen timestamp
	executorHeartbeatDate1 := now.Add(-16 * time.Second) // 16 seconds ago
	executorHeartbeatDate2 := now.Add(-1 * time.Hour)    // 1 hour ago
	executorHeartbeatDate3 := now.Add(-24 * time.Hour)   // 1 day ago

	// active executors last seen timestamp
	executorHeartbeatDate4 := now.Add(12 * time.Second) // 12 seconds ago
	executorHeartbeatDate5 := now.Add(3 * time.Second)  // 3 seconds ago

	// Create 5 executor_heartbeats
	// 2 are active (sent an heartbeat within last 15 seconds) while the remaining are inactive
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO executor_heartbeats
			(id, hostname, queue_name,os,architecture,docker_version,executor_version,git_version,ignite_version,src_cli_version,first_seen_at,last_seen_at)
		VALUES
			-- inactive executors
			(83505,'test-hostname-1.0','batches','darwin','arm64','20.10.12','0.0.0+dev','2.35.1','','dev','2022-04-20 17:09:18.010637+02',$1::timestamp),
			(83595,'test-hostname-2.0','batches','darwin','arm64','20.10.12','0.0.0+dev','2.35.1','','dev','2022-04-20 17:16:51.252115+02',$2::timestamp),
			(83603,'test-hostname-3.0','batches','darwin','arm64','20.10.12','0.0.0+dev','2.35.1','','dev','2022-04-20 17:18:08.288158+02', $3::timestamp),

			-- active executors
			(8450, 'test-hostname-1.1', 'batches', 'darwin', 'arm64', '20.10.12', '0.0.0+dev','2.35.1','','dev','2022-04-20 17:09:18.010637+02', $4::timestamp),
			(8451, 'test-hostname-4.0', 'batches', 'darwin', 'arm64', '20.10.12', '0.0.0+dev','2.35.1','','dev','2022-04-20 17:09:18.010637+02', $5::timestamp)
	`, executorHeartbeatDate1, executorHeartbeatDate2, executorHeartbeatDate3, executorHeartbeatDate4, executorHeartbeatDate5)
	if err != nil {
		t.Fatal(err)
	}

	batchChangeID := 1

	// Create different changeset jobs, consisting of the following job types
	// 2 published, 2 comment, 1 closed, 1 merged, 1 detached, 1 reenqueued
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO changeset_jobs
			(id, bulk_group, user_id, batch_change_id, changeset_id, job_type, payload, state, failure_message, started_at, finished_at, process_after, num_resets, num_failures, execution_logs, created_at, updated_at, worker_hostname, last_heartbeat_at, queued_at)
		VALUES
			-- publish jobs
			(1, '2dT7VN2BN6U', $1, $2, $3, 'publish', '{"draft":true}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostname-1.0', NULL, NULL),
			(2, '2dT7VN2BN7U', $1, $2, $4, 'publish', '{"draft":true}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostname-1.0', NULL, NULL),

			-- comment jobs
			(3, '2dT7VN2BN8U', $1, $2, $5, 'commentatore', '{"message":"hold"}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostname-1.0', NULL, NULL),
			(4, '2dT7VN2BN9U', $1, $2, $6, 'commentatore', '{"message":"hold"}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostname-1.0', NULL, NULL),

			-- close jobs
			(5, '3dT7VN2BN6U', $1, $2, $7, 'close', '{"draft":true}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostname-1.0', NULL, NULL),

			-- merge jobs
			(6, '3dT7VN2BN7U', $1, $2, $3, 'merge', '{"draft":true}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostname-1.0', NULL, NULL),

			-- detached jobs
			(7, '3dT7VN2BN8U', $1, $2, $5, 'detach', '{"draft":true}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostname-1.0', NULL, NULL),

			-- reenqueued jobs
			(8, '3dT7VN2BN3U', $1, $2, $6, 'reenqueue', '{"draft":true}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostname-1.0', NULL, NULL)
	`, user.ID, batchChangeID, changesetIDOne, changesetIDTwo, changesetIDFour, changesetIDFive, changesetIDSix)
	if err != nil {
		t.Fatal(err)
	}

	have, err := GetBatchChangesUsageStatistics(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	currentYear, currentMonth, _ := now.Date()
	pastYear, pastMonth, _ := lastMonthCreationDate.Date()
	pastYear2, pastMonth2, _ := twoMonthsAgoCreationDate.Date()

	want := &types.BatchChangesUsageStatistics{
		ViewBatchChangeApplyPageCount:               2,
		ViewBatchChangeDetailsPageAfterCreateCount:  2,
		ViewBatchChangeDetailsPageAfterUpdateCount:  2,
		BatchChangesCount:                           3,
		BatchChangesClosedCount:                     1,
		PublishedChangesetsUnpublishedCount:         2,
		PublishedChangesetsCount:                    8,
		PublishedChangesetsDiffStatAddedSum:         40,
		PublishedChangesetsDiffStatDeletedSum:       44,
		PublishedChangesetsMergedCount:              2,
		PublishedChangesetsMergedDiffStatAddedSum:   16,
		PublishedChangesetsMergedDiffStatDeletedSum: 12,
		ImportedChangesetsCount:                     2,
		ImportedChangesetsMergedCount:               1,
		BatchSpecsCreatedCount:                      4,
		ChangesetSpecsCreatedCount:                  4,
		CurrentMonthContributorsCount:               2,
		CurrentMonthUsersCount:                      2,
		BatchChangesCohorts: []*types.BatchChangesCohort{
			{
				Week:                     batchChangeCreationDate1.Truncate(24 * 7 * time.Hour).Format("2006-01-02"),
				BatchChangesOpen:         1,
				ChangesetsImported:       1,
				ChangesetsPublished:      3,
				ChangesetsPublishedOpen:  2,
				ChangesetsPublishedDraft: 1,
			},
			{
				Week:                      batchChangeCreationDate2.Truncate(24 * 7 * time.Hour).Format("2006-01-02"),
				BatchChangesClosed:        1,
				ChangesetsImported:        1,
				ChangesetsUnpublished:     2,
				ChangesetsPublished:       3,
				ChangesetsPublishedMerged: 2,
				ChangesetsPublishedClosed: 1,
			},
			// batch change 3 should be ignored because it's too old
		},
		ActiveExecutorsCount: 2,
		BulkOperationsCount: []*types.BulkOperationsCount{
			{Name: "close", Count: 1},
			{Name: "comment", Count: 2},
			{Name: "detach", Count: 1},
			{Name: "merge", Count: 1},
			{Name: "publish", Count: 2},
			{Name: "reenqueue", Count: 1},
		},
		ChangesetDistribution: []*types.ChangesetDistribution{
			{Source: "local", Range: "0-9 changesets", BatchChangesCount: 2},
			{Source: "executor", Range: "0-9 changesets", BatchChangesCount: 1},
		},
		BatchChangeStatsBySource: []*types.BatchChangeStatsBySource{
			{
				Source:                   "local",
				PublishedChangesetsCount: 8,
				BatchChangesCount:        2,
			},
			{
				Source:                   "executor",
				PublishedChangesetsCount: 2,
				BatchChangesCount:        1,
			},
		},
		MonthlyBatchChangesExecutorUsage: []*types.MonthlyBatchChangesExecutorUsage{
			{Month: fmt.Sprintf("%d-%02d-01T00:00:00Z", pastYear2, pastMonth2), Count: 1, Minutes: 0},
			{Month: fmt.Sprintf("%d-%02d-01T00:00:00Z", pastYear, pastMonth), Count: 2, Minutes: 10},
			{Month: fmt.Sprintf("%d-%02d-01T00:00:00Z", currentYear, currentMonth), Count: 1, Minutes: 30},
		},
		WeeklyBulkOperationStats: []*types.WeeklyBulkOperationStats{
			{
				Week:          "2022-03-21T00:00:00Z",
				Count:         1,
				BulkOperation: "close",
			},
			{
				Week:          "2022-03-21T00:00:00Z",
				Count:         2,
				BulkOperation: "comment",
			},
			{
				Week:          "2022-03-21T00:00:00Z",
				Count:         1,
				BulkOperation: "detach",
			},
			{
				Week:          "2022-03-21T00:00:00Z",
				Count:         1,
				BulkOperation: "merge",
			},
			{
				Week:          "2022-03-21T00:00:00Z",
				Count:         2,
				BulkOperation: "publish",
			},
			{
				Week:          "2022-03-21T00:00:00Z",
				Count:         1,
				BulkOperation: "reenqueue",
			},
		},
	}

	sort.Slice(have.BulkOperationsCount, func(i, j int) bool {
		return have.BulkOperationsCount[i].Name < have.BulkOperationsCount[j].Name
	})

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}
