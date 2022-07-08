CREATE SEQUENCE create_webhook_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE create_webhook_jobs (
    id integer DEFAULT nextval('create_webhook_jobs_id_seq'::regclass) NOT NULL,
    state text DEFAULT 'queued'::text NOT NULL,
    failure_message text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    process_after timestamp with time zone,
    num_resets integer DEFAULT 0 NOT NULL,
    num_failures integer DEFAULT 0 NOT NULL,
    repo_id bigint,
    repo_name text,
    queued_at timestamp with time zone,
    execution_logs json[],
    last_heartbeat_at timestamp with time zone,
    worker_hostname text DEFAULT ''::text NOT NULL
);

ALTER TABLE ONLY create_webhook_jobs
    ADD CONSTRAINT create_webhook_jobs_fk FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE CASCADE;
