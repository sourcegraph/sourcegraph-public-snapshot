CREATE TABLE IF NOT EXISTS executor_job_tokens
(
    id           SERIAL PRIMARY KEY,
    value_sha256 bytea                                  NOT NULL,
    job_id       BIGINT                                 NOT NULL,
    queue        TEXT                                   NOT NULL,
    repo_id      BIGINT                                 NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

ALTER TABLE executor_job_tokens
    DROP CONSTRAINT IF EXISTS executor_job_tokens_value_sha256_key;
ALTER TABLE ONLY executor_job_tokens
    ADD CONSTRAINT executor_job_tokens_value_sha256_key UNIQUE (value_sha256);

ALTER TABLE executor_job_tokens
    DROP CONSTRAINT IF EXISTS executor_job_tokens_job_id_queue_repo_id_key;
ALTER TABLE ONLY executor_job_tokens
    ADD CONSTRAINT executor_job_tokens_job_id_queue_repo_id_key UNIQUE (job_id, queue, repo_id);
