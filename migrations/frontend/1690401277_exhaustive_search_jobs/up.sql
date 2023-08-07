CREATE TABLE IF NOT EXISTS exhaustive_search_jobs
(
    id                SERIAL PRIMARY KEY,
    state             text                     DEFAULT 'queued'::text,
    initiator_id      integer                                NOT NULL,
    query             text                                   NOT NULL,
    failure_message   text,
    started_at        timestamp with time zone,
    finished_at       timestamp with time zone,
    process_after     timestamp with time zone,
    num_resets        integer                  DEFAULT 0     NOT NULL,
    num_failures      integer                  DEFAULT 0     NOT NULL,
    last_heartbeat_at timestamp with time zone,
    execution_logs    json[],
    worker_hostname   text                                   not null default '',
    cancel            boolean                                not null default false,
    created_at        timestamp with time zone DEFAULT now() NOT NULL,
    updated_at        timestamp with time zone DEFAULT now() NOT NULL,
    queued_at         timestamp with time zone DEFAULT now()
);

ALTER TABLE exhaustive_search_jobs
    DROP CONSTRAINT IF EXISTS exhaustive_search_jobs_initiator_id_fkey,
    ADD CONSTRAINT exhaustive_search_jobs_initiator_id_fkey
        FOREIGN KEY (initiator_id)
            REFERENCES users (id)
            ON DELETE CASCADE
            ON UPDATE CASCADE
            DEFERRABLE;
