-- Perform migration here.
--
-- See /migrations/README.md. Highlights:
--  * Make migrations idempotent (use IF EXISTS)
--  * Make migrations backwards-compatible (old readers/writers must continue to work)
--  * If you are using CREATE INDEX CONCURRENTLY, then make sure that only one statement
--    is defined per file, and that each such statement is NOT wrapped in a transaction.
--    Each such migration must also declare "createIndexConcurrently: true" in their
--    associated metadata.yaml file.
--  * If you are modifying Postgres extensions, you must also declare "privileged: true"
--    in the associated metadata.yaml file.

CREATE TABLE IF NOT EXISTS executor_job_tokens
(
    id           SERIAL PRIMARY KEY,
    value_sha256 bytea                                  NOT NULL,
    job_id       BIGINT                                 NOT NULL,
    queue        TEXT                                   NOT NULL,
    repo         TEXT                                   NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

ALTER TABLE ONLY executor_job_tokens
    ADD CONSTRAINT executor_job_tokens_value_sha256_key UNIQUE (value_sha256);

ALTER TABLE ONLY executor_job_tokens
    ADD CONSTRAINT executor_job_tokens_job_id_queue_repo_key UNIQUE (job_id, queue, repo);
