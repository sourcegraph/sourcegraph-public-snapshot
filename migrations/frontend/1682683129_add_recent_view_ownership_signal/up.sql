CREATE TABLE IF NOT EXISTS own_aggregate_recent_view
(
    id                  SERIAL PRIMARY KEY,
    viewer_id           INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE DEFERRABLE,
    viewed_file_path_id INTEGER NOT NULL REFERENCES repo_paths (id),
    views_count         INTEGER DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS own_aggregate_recent_view_viewer
    ON own_aggregate_recent_view
        USING btree (viewed_file_path_id, viewer_id);

COMMENT ON TABLE own_aggregate_recent_view
    IS 'One entry contains a number of views of a single file by a given viewer.';

CREATE TABLE IF NOT EXISTS event_logs_scrape_state_own
(
    id          SERIAL
        CONSTRAINT event_logs_scrape_state_own_pk
            PRIMARY KEY,
    bookmark_id INT NOT NULL,
    job_type    INT NOT NULL
);

COMMENT ON TABLE event_logs_scrape_state_own IS 'Contains state for own jobs that scrape events if enabled.';
COMMENT ON COLUMN event_logs_scrape_state_own.bookmark_id IS 'Bookmarks the maximum most recent successful event_logs.id that was scraped';
