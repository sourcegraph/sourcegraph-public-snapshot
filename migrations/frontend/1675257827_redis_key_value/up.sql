-- Perform migration here.

CREATE TABLE IF NOT EXISTS redis_key_value (
    namespace TEXT NOT NULL,
    key TEXT NOT NULL,
    value BYTEA NOT NULL,
    PRIMARY KEY (namespace, key) INCLUDE (value)
);
