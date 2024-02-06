
CREATE TABLE syntactic_scip_indexes (
    id bigint NOT NULL,
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
    outfile text NOT NULL,
    execution_logs json[],
    commit_last_checked_at timestamp with time zone,
    worker_hostname text DEFAULT ''::text NOT NULL,
    last_heartbeat_at timestamp with time zone,
    cancel boolean DEFAULT false NOT NULL,
    should_reindex boolean DEFAULT false NOT NULL,
    enqueuer_user_id integer DEFAULT 0 NOT NULL,
    CONSTRAINT syntactic_scip_indexes_commit_valid_chars CHECK ((commit ~ '^[a-z0-9]{40}$'::text))
);

CREATE SEQUENCE syntactic_scip_index_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE syntactic_scip_index_id_seq OWNED BY syntactic_scip_indexes.id;

COMMENT ON TABLE syntactic_scip_indexes IS 'Stores metadata about a code intel syntactic index job.';

COMMENT ON COLUMN syntactic_scip_indexes.commit IS 'A 40-char revhash. Note that this commit may not be resolvable in the future.';

COMMENT ON COLUMN syntactic_scip_indexes.outfile IS 'The path to the index file produced by the index command relative to the working directory.';

COMMENT ON COLUMN syntactic_scip_indexes.execution_logs IS 'An array of [log entries](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/internal/workerutil/store.go#L48:6) (encoded as JSON) from the most recent execution.';

COMMENT ON COLUMN syntactic_scip_indexes.enqueuer_user_id IS 'ID of the user who scheduled this index. Records with a non-NULL user ID are prioritised over the rest';


DROP VIEW IF EXISTS syntactic_scip_indexes_with_repository_name;

CREATE VIEW syntactic_scip_indexes_with_repository_name AS
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
        u.outfile,
        u.execution_logs,
        u.should_reindex,
        u.enqueuer_user_id,
        r.name AS repository_name
    FROM (syntactic_scip_indexes u
        JOIN repo r ON ((r.id = u.repository_id)))
    WHERE (r.deleted_at IS NULL);
