BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm
-- CREATE TABLE IF NOT EXISTS saved_queries (
--
-- );

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
