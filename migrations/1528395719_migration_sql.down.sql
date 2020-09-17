BEGIN;
CREATE TABLE IF NOT EXISTS  saved_queries
(
    query text not null,
    last_executed timestamp with time zone not null,
    latest_result timestamp with time zone not null,
    exec_duration_ns bigint not null
);

alter table saved_queries owner to sourcegraph;

create unique index saved_queries_query_unique
    on saved_queries (query);

COMMIT;
