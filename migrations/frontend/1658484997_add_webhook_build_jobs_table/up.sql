CREATE SEQUENCE IF NOT EXISTS webhook_build_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

 CREATE TABLE IF NOT EXISTS webhook_build_jobs (
    repo_id integer,
    repo_name text,
    extsvc_kind text,
    queued_at timestamp with time zone DEFAULT now(),
    id integer DEFAULT nextval('webhook_build_jobs_id_seq'::regclass) NOT NULL,
    state text DEFAULT 'queued'::text NOT NULL,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    execution_logs json[],
    last_heartbeat_at timestamp with time zone,
    worker_hostname text DEFAULT ''::text NOT NULL
 );

 CREATE INDEX IF NOT EXISTS webhook_build_jobs_queued_at_idx ON webhook_build_jobs USING btree (queued_at);
