CREATE TABLE saved_queries (
    "query" text NOT NULL,
    "last_executed" TIMESTAMP WITH TIME ZONE NOT NULL,
    "latest_result" TIMESTAMP WITH TIME ZONE NOT NULL,
    "exec_duration_ns" bigint NOT NULL
);
CREATE UNIQUE INDEX "saved_queries_query_unique" ON saved_queries(query);
COMMIT;