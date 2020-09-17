BEGIN;
CREATE TABLE IF NOT EXISTS saved_queries
(
    query TEXT NOT NULL,
    last_executed TIMESTAMP WITH TIME ZONE NOT NULL,
    latest_result TIMESTAMP WITH TIME ZONE NOT NULL,
    exec_duration_ns BIGINT NOT NULL
);

ALTER TABLE saved_queries OWNER TO sourcegraph;

CREATE UNIQUE INDEX saved_queries_query_unique
    ON saved_queries (query);

COMMIT;
