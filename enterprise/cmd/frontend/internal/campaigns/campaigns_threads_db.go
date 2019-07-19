package campaigns

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// dbCampaignThread represents a thread's inclusion in a campaign.
type dbCampaignThread struct {
	Campaign int64 // the ID of the campaign
	Thread   int64 // the ID of the thread
}

type dbCampaignsThreads struct{}

// AddThreadsToCampaign adds threads to the campaign.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to modify the campaign and read
// the objects.
func (dbCampaignsThreads) AddThreadsToCampaign(ctx context.Context, campaign int64, threads []int64) error {
	if mocks.campaignsThreads.AddThreadsToCampaign != nil {
		return mocks.campaignsThreads.AddThreadsToCampaign(campaign, threads)
	}

	_, err := dbconn.Global.ExecContext(ctx,
		`
WITH insert_values AS (
  SELECT $1::bigint AS campaign_id, unnest($2::bigint[]) AS thread_id
)
INSERT INTO campaigns_threads(campaign_id, thread_id) SELECT * FROM insert_values
`,
		campaign, pq.Array(threads),
	)
	return err
}

// RemoveThreadsFromCampaign removes the specified threads from the thread.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to modify the thread and the
// threads.
func (dbCampaignsThreads) RemoveThreadsFromCampaign(ctx context.Context, campaign int64, threads []int64) error {
	if mocks.campaignsThreads.RemoveThreadsFromCampaign != nil {
		return mocks.campaignsThreads.RemoveThreadsFromCampaign(campaign, threads)
	}

	_, err := dbconn.Global.ExecContext(ctx,
		`
WITH delete_values AS (
  SELECT $1::bigint AS campaign_id, unnest($2::bigint[]) AS thread_id
)
DELETE FROM campaigns_threads o USING delete_values d WHERE o.campaign_id=d.campaign_id AND o.thread_id=d.thread_id
`,
		campaign, pq.Array(threads),
	)
	return err
}

// dbCampaignsThreadsListOptions contains options for listing threads.
type dbCampaignsThreadsListOptions struct {
	CampaignID int64 // only list threads for this campaign
	ThreadID   int64 // only list campaigns for this thread
	*db.LimitOffset
}

func (o dbCampaignsThreadsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.CampaignID != 0 {
		conds = append(conds, sqlf.Sprintf("campaign_id=%d", o.CampaignID))
	}
	if o.ThreadID != 0 {
		conds = append(conds, sqlf.Sprintf("thread_id=%d", o.ThreadID))
	}
	return conds
}

// List lists all campaign-thread associations that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s dbCampaignsThreads) List(ctx context.Context, opt dbCampaignsThreadsListOptions) ([]*dbCampaignThread, error) {
	if mocks.campaignsThreads.List != nil {
		return mocks.campaignsThreads.List(opt)
	}

	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s dbCampaignsThreads) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbCampaignThread, error) {
	q := sqlf.Sprintf(`
SELECT campaign_id, thread_id FROM campaigns_threads
WHERE (%s) AND thread_id IS NOT NULL
ORDER BY campaign_id ASC, thread_id ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)
	return s.query(ctx, q)
}

func (dbCampaignsThreads) query(ctx context.Context, query *sqlf.Query) ([]*dbCampaignThread, error) {
	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbCampaignThread
	for rows.Next() {
		var t dbCampaignThread
		if err := rows.Scan(&t.Campaign, &t.Thread); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, nil
}

// mockCampaignsThreads mocks the campaigns-threads-related DB operations.
type mockCampaignsThreads struct {
	AddThreadsToCampaign      func(thread int64, threads []int64) error
	RemoveThreadsFromCampaign func(thread int64, threads []int64) error
	List                      func(dbCampaignsThreadsListOptions) ([]*dbCampaignThread, error)
}
