CREATE TABLE IF NOT EXISTS security_event_logs_export_allowlist
(
    id         SERIAL PRIMARY KEY,
    event_name TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS security_event_logs_export_allowlist_event_name_idx ON security_event_logs_export_allowlist (event_name);

COMMENT ON TABLE security_event_logs_export_allowlist IS 'An allowlist of security events that are approved for export if the scraping job is enabled';
COMMENT ON COLUMN security_event_logs_export_allowlist.event_name IS 'Name of the security event that corresponds to security_event_logs.name';
