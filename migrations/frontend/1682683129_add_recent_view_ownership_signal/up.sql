CREATE TABLE IF NOT EXISTS own_aggregate_recent_view
(
    id                  SERIAL PRIMARY KEY,
    viewer_email        TEXT    NOT NULL,
    viewer_name         TEXT    NOT NULL,
    viewed_file_path_id INTEGER NOT NULL REFERENCES repo_paths (id),
    views_count         INTEGER DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS own_aggregate_recent_view_viewer
    ON own_aggregate_recent_view
        USING btree (viewed_file_path_id, viewer_email, viewer_name);

COMMENT ON TABLE own_aggregate_recent_view
    IS 'One entry contains a number of views of a single file by a given viewer.';
