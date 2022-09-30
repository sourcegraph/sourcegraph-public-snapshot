-- Perform migration here.

CREATE TABLE IF NOT EXISTS aggregated_user_statistics (
    user_id BIGINT NOT NULL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    user_last_active_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    user_events_count BIGINT DEFAULT NULL,
    CONSTRAINT aggregated_user_statistics_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
