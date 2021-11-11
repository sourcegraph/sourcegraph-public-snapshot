BEGIN;

CREATE TABLE IF NOT EXISTS executor_heartbeats (
    id SERIAL PRIMARY KEY,
    hostname TEXT NOT NULL UNIQUE,
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE executor_heartbeats IS 'Tracks the most recent activity of executors attached to this Sourcegraph instance.';
COMMENT ON COLUMN executor_heartbeats.hostname IS 'The uniquely identifying name of the executor.';
COMMENT ON COLUMN executor_heartbeats.last_seen_at IS 'The last time a heartbeat from the executor was received.';

COMMIT;
