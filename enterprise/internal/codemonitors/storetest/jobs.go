package storetest

import (
	"context"

	"github.com/keegancsmith/sqlf"
)

type JobTable int

const (
	TriggerJobs JobTable = iota
	ActionJobs
)

type JobState int

const (
	Queued JobState = iota
	Processing
	Completed
	Errored
	Failed
)

const setStatusFmtStr = `
UPDATE %s
SET state = %s,
    started_at = %s,
    finished_at = %s
WHERE id = %s;
`

// quote wraps the given string in a *sqlf.Query so that it is not passed to the database
// as a parameter. It is necessary to quote things such as table names, columns, and other
// expressions that are not simple values.
func quote(s string) *sqlf.Query {
	return sqlf.Sprintf(s)
}

func (s *TestStore) SetJobStatus(ctx context.Context, table JobTable, state JobState, id int) error {
	st := []string{"queued", "processing", "completed", "errored", "failed"}[state]
	t := []string{"cm_trigger_jobs", "cm_action_jobs"}[table]
	return s.Exec(ctx, sqlf.Sprintf(setStatusFmtStr, quote(t), st, s.Now(), s.Now(), id))
}
