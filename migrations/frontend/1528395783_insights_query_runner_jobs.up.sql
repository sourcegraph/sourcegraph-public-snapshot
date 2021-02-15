BEGIN;

CREATE TABLE IF NOT EXISTS insights_query_runner_jobs(
    id SERIAL PRIMARY KEY,
    series_id text NOT NULL,
    search_query text NOT NULL,
    state           text default 'queued',
    failure_message text,
    started_at      timestamptz,
    finished_at     timestamptz,
    process_after   timestamptz,
    num_resets      int4 NOT NULL default 0,
    num_failures    int4 NOT NULL default 0,
    execution_logs json[]
);
CREATE INDEX insights_query_runner_jobs_state_btree ON insights_query_runner_jobs USING btree (state);

COMMENT ON TABLE insights_query_runner_jobs IS 'See [enterprise/internal/insights/background/queryrunner/worker.go:Job](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:enterprise/internal/insights/background/queryrunner/worker.go+type+Job&patternType=literal)';

COMMIT;
