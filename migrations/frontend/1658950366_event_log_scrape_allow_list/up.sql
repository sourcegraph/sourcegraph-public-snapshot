CREATE TABLE IF NOT EXISTS event_logs_export_allowlist
(
    id         SERIAL PRIMARY KEY,
    event_name TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS event_logs_export_allowlist_event_name_idx ON event_logs_export_allowlist (event_name);

COMMENT ON TABLE event_logs_export_allowlist IS 'An allowlist of events that are approved for export if the scraping job is enabled';
COMMENT ON COLUMN event_logs_export_allowlist.event_name IS 'Name of the event that corresponds to event_logs.name';
