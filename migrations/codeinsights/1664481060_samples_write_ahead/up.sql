CREATE TABLE IF NOT EXISTS samples_write_ahead
(
    id        SERIAL
        CONSTRAINT samples_write_ahead_pk PRIMARY KEY,
    series_id INT              NOT NULL,
    repo_id   INT              NOT NULL,
    capture   TEXT,
    time      INT              NOT NULL,
    value     DOUBLE PRECISION NOT NULL
);

CREATE INDEX IF NOT EXISTS samples_write_ahead_compound_idx on samples_write_ahead(series_id, repo_id, capture);
