CREATE TABLE IF NOT EXISTS security_event_logs_scrape_state
(
    id                     SERIAL
        CONSTRAINT security_event_logs_scrape_state_pk
            PRIMARY KEY,
    bookmark_id INT NOT NULL
);

COMMENT ON TABLE security_event_logs_scrape_state IS 'Contains state for the periodic telemetry job that scrapes security events if enabled.';
COMMENT ON COLUMN security_event_logs_scrape_state.bookmark_id IS 'Bookmarks the maximum most recent successful security_event_logs.id that was scraped';
