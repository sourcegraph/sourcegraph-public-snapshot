CREATE TABLE IF NOT EXISTS  telemetry_events_export_queue (
  id TEXT PRIMARY KEY,
  timestamp TIMESTAMPTZ NOT NULL,
  payload_pb BYTEA NOT NULL,
  exported_at TIMESTAMPTZ DEFAULT NULL
);
