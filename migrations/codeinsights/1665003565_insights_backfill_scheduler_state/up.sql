CREATE TABLE IF NOT EXISTS repo_iterator
(
    id               SERIAL
        CONSTRAINT repo_iterator_pk PRIMARY KEY,
    created_at       TIMESTAMP          DEFAULT NOW(),
    started_at       TIMESTAMP,
    completed_at     TIMESTAMP,
    last_updated_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    runtime_duration BIGINT NOT NULL DEFAULT 0,
    percent_complete DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_count            INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    repos            INT[],
    repo_cursor      INT                DEFAULT 0
);

CREATE TABLE IF NOT EXISTS repo_iterator_errors
(
    id               SERIAL
        CONSTRAINT repo_iterator_errors_pk PRIMARY KEY,
    repo_iterator_id INT    NOT NULL,
    repo_id          INT    NOT NULL,
    error_message    TEXT[] NOT NULL,
    failure_count    INT DEFAULT 1,

    CONSTRAINT repo_iterator_fk FOREIGN KEY (repo_iterator_id) REFERENCES repo_iterator (id)
);

CREATE INDEX IF NOT EXISTS repo_iterator_errors_fk_idx
    ON repo_iterator_errors (repo_iterator_id);
