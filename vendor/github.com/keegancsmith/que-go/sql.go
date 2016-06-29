// Copyright (c) 2013 Chris Hanks
//
// MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package que

// Thanks to RhodiumToad in #postgresql for help with the job lock CTE.
const (
	sqlLockJob = `
WITH RECURSIVE jobs AS (
  SELECT (j).*, pg_try_advisory_xact_lock((j).job_id) AS locked
  FROM (
    SELECT j
    FROM que_jobs AS j
    WHERE queue = $1::text
    AND run_at <= now()
    ORDER BY priority, run_at, job_id
    LIMIT 1
  ) AS t1
  UNION ALL (
    SELECT (j).*, pg_try_advisory_xact_lock((j).job_id) AS locked
    FROM (
      SELECT (
        SELECT j
        FROM que_jobs AS j
        WHERE queue = $1::text
        AND run_at <= now()
        AND (priority, run_at, job_id) > (jobs.priority, jobs.run_at, jobs.job_id)
        ORDER BY priority, run_at, job_id
        LIMIT 1
      ) AS j
      FROM jobs
      WHERE jobs.job_id IS NOT NULL
      LIMIT 1
    ) AS t1
  )
)
SELECT queue, priority, run_at, job_id, job_class, args, error_count
FROM jobs
WHERE locked
LIMIT 1
`

	sqlUnlockJob = `
SELECT pg_advisory_unlock($1)
`

	sqlCheckJob = `
SELECT true AS exists
FROM   que_jobs
WHERE  queue    = $1::text
AND    priority = $2::smallint
AND    run_at   = $3::timestamptz
AND    job_id   = $4::bigint
`

	sqlSetError = `
UPDATE que_jobs
SET error_count = $1::integer,
    run_at      = now() + $2::bigint * '1 second'::interval,
    last_error  = $3::text
WHERE queue     = $4::text
AND   priority  = $5::smallint
AND   run_at    = $6::timestamptz
AND   job_id    = $7::bigint
`

	sqlInsertJob = `
INSERT INTO que_jobs
(queue, priority, run_at, job_class, args)
VALUES
(coalesce($1::text, ''::text), coalesce($2::smallint, 100::smallint), coalesce($3::timestamptz, now()::timestamptz), $4::text, coalesce($5::json, '[]'::json))
`

	sqlDeleteJob = `
DELETE FROM que_jobs
WHERE queue    = $1::text
AND   priority = $2::smallint
AND   run_at   = $3::timestamptz
AND   job_id   = $4::bigint
`

	sqlJobStats = `
SELECT queue,
       job_class,
       count(*)                    AS count,
       count(locks.job_id)         AS count_working,
       sum((error_count > 0)::int) AS count_errored,
       max(error_count)            AS highest_error_count,
       min(run_at)                 AS oldest_run_at
FROM que_jobs
LEFT JOIN (
  SELECT (classid::bigint << 32) + objid::bigint AS job_id
  FROM pg_locks
  WHERE locktype = 'advisory'
) locks USING (job_id)
GROUP BY queue, job_class
ORDER BY count(*) DESC
`

	sqlWorkerStates = `
SELECT que_jobs.*,
       pg.pid          AS pg_backend_pid,
       pg.state        AS pg_state,
       pg.state_change AS pg_state_changed_at,
       pg.query        AS pg_last_query,
       pg.query_start  AS pg_last_query_started_at,
       pg.xact_start   AS pg_transaction_started_at,
       pg.waiting      AS pg_waiting_on_lock
FROM que_jobs
JOIN (
  SELECT (classid::bigint << 32) + objid::bigint AS job_id, pg_stat_activity.*
  FROM pg_locks
  JOIN pg_stat_activity USING (pid)
  WHERE locktype = 'advisory'
) pg USING (job_id)
`
)
