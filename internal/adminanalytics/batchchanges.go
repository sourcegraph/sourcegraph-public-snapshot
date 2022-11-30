package adminanalytics

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type BatchChanges struct {
	Ctx       context.Context
	DateRange string
	Grouping  string
	DB        database.DB
	Cache     bool
}

var changesetsCreatedNodesQuery = `
	SELECT
		%s AS date,
		COUNT(DISTINCT changesets.id) AS count,
		COUNT(DISTINCT batch_changes.creator_id) AS unique_users,
		COUNT(DISTINCT batch_changes.creator_id) AS registered_users
	FROM
		changesets
		INNER JOIN batch_changes ON batch_changes.id = changesets.owned_by_batch_change_id
	WHERE changesets.created_at %s AND changesets.publication_state = 'PUBLISHED'
	GROUP BY date
`

var changesetsCreatedSummaryQuery = `
	SELECT
		COUNT(DISTINCT changesets.id) AS total_count,
		COUNT(DISTINCT batch_changes.creator_id) AS total_unique_users,
		COUNT(DISTINCT batch_changes.creator_id) AS total_registered_users
	FROM
		changesets
		INNER JOIN batch_changes ON batch_changes.id = changesets.owned_by_batch_change_id
	WHERE changesets.created_at %s AND changesets.publication_state = 'PUBLISHED'
`

func (s *BatchChanges) ChangesetsCreated() (*AnalyticsFetcher, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(s.DateRange, s.Grouping, "changesets.created_at")
	if err != nil {
		return nil, err
	}

	nodesQuery := sqlf.Sprintf(changesetsCreatedNodesQuery, dateTruncExp, dateBetweenCond)
	summaryQuery := sqlf.Sprintf(changesetsCreatedSummaryQuery, dateBetweenCond)

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		grouping:     s.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "BatchChanges:ChangesetsCreated",
		cache:        s.Cache,
	}, nil
}

var changesetsMergedNodesQuery = `
	SELECT
		%s AS date,
		COUNT(DISTINCT changesets.id) AS count,
		COUNT(DISTINCT batch_changes.creator_id) AS unique_users,
		COUNT(DISTINCT batch_changes.creator_id) AS registered_users
	FROM
		changeset_events
		INNER JOIN changesets ON changesets.id = changeset_events.changeset_id
		INNER JOIN batch_changes ON batch_changes.id = changesets.owned_by_batch_change_id
	WHERE changeset_events.created_at %s AND changeset_events.kind IN (%s)
	GROUP BY date
`

var changesetsMergedSummaryQuery = `
	SELECT
		COUNT(DISTINCT changesets.id) AS total_count,
		COUNT(DISTINCT batch_changes.creator_id) AS total_unique_users,
		COUNT(DISTINCT batch_changes.creator_id) AS total_registered_users
	FROM
		changeset_events
		INNER JOIN changesets ON changesets.id = changeset_events.changeset_id
		INNER JOIN batch_changes ON batch_changes.id = changesets.owned_by_batch_change_id
	WHERE changeset_events.created_at %s AND changeset_events.kind IN (%s)
`

var mergeEventKinds = sqlf.Join([]*sqlf.Query{
	sqlf.Sprintf("'github:merged'"),
	sqlf.Sprintf("'bitbucketserver:merged'"),
	sqlf.Sprintf("'gitlab:merged'"),
	sqlf.Sprintf("'bitbucketcloud:pullrequest:fulfilled'"),
}, ",")

func (s *BatchChanges) ChangesetsMerged() (*AnalyticsFetcher, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(s.DateRange, s.Grouping, "changesets.created_at")
	if err != nil {
		return nil, err
	}

	nodesQuery := sqlf.Sprintf(changesetsMergedNodesQuery, dateTruncExp, dateBetweenCond, mergeEventKinds)
	summaryQuery := sqlf.Sprintf(changesetsMergedSummaryQuery, dateBetweenCond, mergeEventKinds)

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		grouping:     s.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "BatchChanges:ChangesetsMerged",
		cache:        s.Cache,
	}, nil
}

func (s *BatchChanges) CacheAll(ctx context.Context) error {
	fetcherBuilders := []func() (*AnalyticsFetcher, error){s.ChangesetsCreated, s.ChangesetsMerged}
	for _, buildFetcher := range fetcherBuilders {
		fetcher, err := buildFetcher()
		if err != nil {
			return err
		}

		if _, err := fetcher.Nodes(ctx); err != nil {
			return err
		}

		if _, err := fetcher.Summary(ctx); err != nil {
			return err
		}
	}
	return nil
}
