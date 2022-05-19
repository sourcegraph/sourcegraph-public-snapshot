CREATE TABLE IF NOT EXISTS last_lockfile_scan (
    repository_id integer NOT NULL,
    last_lockfile_scan_at timestamp with time zone NOT NULL
);

ALTER TABLE ONLY last_lockfile_scan
    ADD CONSTRAINT last_lockfile_scan_pkey PRIMARY KEY (repository_id);

COMMENT ON TABLE last_lockfile_scan IS 'Tracks the last time repository was checked for lockfile indexing.';

COMMENT ON COLUMN last_lockfile_scan.last_lockfile_scan_at IS 'The last time this repository was considered for lockfile indexing.';
