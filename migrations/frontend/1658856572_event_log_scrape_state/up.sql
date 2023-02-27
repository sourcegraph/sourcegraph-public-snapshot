CREATE TABLE IF NOT EXISTS event_logs_scrape_state
(
    id                     SERIAL
        CONSTRAINT event_logs_scrape_state_pk
            PRIMARY KEY,
    bookmark_id INT NOT NULL
);

COMMENT ON TABLE event_logs_scrape_state IS 'Contains state for the periodic telemetry job that scrapes events if enabled.';
COMMENT ON COLUMN event_logs_scrape_state.bookmark_id IS 'Bookmarks the maximum most recent successful event_logs.id that was scraped';
