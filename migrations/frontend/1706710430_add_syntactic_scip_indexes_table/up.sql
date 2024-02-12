
CREATE TABLE IF NOT EXISTS syntactic_scip_indexing_jobs (
    id bigserial NOT NULL PRIMARY KEY,
    commit text NOT NULL,
    queued_at timestamp with time zone DEFAULT now() NOT NULL,
    state text DEFAULT 'queued'::text NOT NULL,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    repository_id integer NOT NULL,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    execution_logs json[],
    commit_last_checked_at timestamp with time zone,
    worker_hostname text DEFAULT ''::text NOT NULL,
    last_heartbeat_at timestamp with time zone,
    cancel boolean DEFAULT false NOT NULL,
    should_reindex boolean DEFAULT false NOT NULL,
    enqueuer_user_id integer DEFAULT 0 NOT NULL,
    CONSTRAINT syntactic_scip_indexing_jobs_commit_valid_chars CHECK ((commit ~ '^[a-f0-9]{40}$'::text))
);


CREATE INDEX IF NOT EXISTS syntactic_scip_indexing_jobs_dequeue_order_idx
    ON syntactic_scip_indexing_jobs
    USING btree (((enqueuer_user_id > 0)) DESC, queued_at DESC, id)
    WHERE ((state = 'queued'::text) OR (state = 'errored'::text));

CREATE INDEX IF NOT EXISTS syntactic_scip_indexing_jobs_queued_at_id ON syntactic_scip_indexing_jobs USING btree (queued_at DESC, id);

CREATE INDEX IF NOT EXISTS syntactic_scip_indexing_jobs_repository_id_commit ON syntactic_scip_indexing_jobs USING btree (repository_id, commit);

CREATE INDEX IF NOT EXISTS syntactic_scip_indexing_jobs_state ON syntactic_scip_indexing_jobs USING btree (state);

COMMENT ON TABLE syntactic_scip_indexing_jobs IS 'Stores metadata about a code intel syntactic index job.';

COMMENT ON COLUMN syntactic_scip_indexing_jobs.commit IS 'A 40-char revhash. Note that this commit may not be resolvable in the future.';

COMMENT ON COLUMN syntactic_scip_indexing_jobs.execution_logs IS 'An array of [log entries](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/internal/workerutil/store.go#L48:6) (encoded as JSON) from the most recent execution.';

COMMENT ON COLUMN syntactic_scip_indexing_jobs.enqueuer_user_id IS 'ID of the user who scheduled this index. Records with a non-NULL user ID are prioritised over the rest';

CREATE OR REPLACE VIEW syntactic_scip_indexing_jobs_with_repository_name AS
    SELECT u.id,
        u.commit,
        u.queued_at,
        u.state,
        u.failure_message,
        u.started_at,
        u.finished_at,
        u.repository_id,
        u.process_after,
        u.num_resets,
        u.num_failures,
        u.execution_logs,
        u.should_reindex,
        u.enqueuer_user_id,
        r.name AS repository_name
    FROM (syntactic_scip_indexing_jobs u
        JOIN repo r ON ((r.id = u.repository_id)))
    WHERE (r.deleted_at IS NULL);
