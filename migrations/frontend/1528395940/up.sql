BEGIN;

CREATE TABLE IF NOT EXISTS executor_heartbeats (
    id SERIAL PRIMARY KEY,
    hostname TEXT NOT NULL UNIQUE,
    queue_name TEXT NOT NULL,
    os TEXT NOT NULL,
    architecture TEXT NOT NULL,
    docker_version TEXT NOT NULL,
    executor_version TEXT NOT NULL,
    git_version TEXT NOT NULL,
    ignite_version TEXT NOT NULL,
    src_cli_version TEXT NOT NULL,
    first_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE executor_heartbeats IS 'Tracks the most recent activity of executors attached to this Sourcegraph instance.';
COMMENT ON COLUMN executor_heartbeats.hostname IS 'The uniquely identifying name of the executor.';
COMMENT ON COLUMN executor_heartbeats.queue_name IS 'The queue name that the executor polls for work.';
COMMENT ON COLUMN executor_heartbeats.os IS 'The operating system running the executor.';
COMMENT ON COLUMN executor_heartbeats.architecture IS 'The machine architure running the executor.';
COMMENT ON COLUMN executor_heartbeats.docker_version IS 'The version of Docker used by the executor.';
COMMENT ON COLUMN executor_heartbeats.executor_version IS 'The version of the executor.';
COMMENT ON COLUMN executor_heartbeats.git_version IS 'The version of Git used by the executor.';
COMMENT ON COLUMN executor_heartbeats.ignite_version IS 'The version of Ignite used by the executor.';
COMMENT ON COLUMN executor_heartbeats.src_cli_version IS 'The version of src-cli used by the executor.';
COMMENT ON COLUMN executor_heartbeats.first_seen_at IS 'The first time a heartbeat from the executor was received.';
COMMENT ON COLUMN executor_heartbeats.last_seen_at IS 'The last time a heartbeat from the executor was received.';

COMMIT;
