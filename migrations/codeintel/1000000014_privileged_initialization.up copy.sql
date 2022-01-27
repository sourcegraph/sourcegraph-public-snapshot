-- This transaction requires SUPERUSER privileges.

BEGIN;

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

COMMENT ON EXTENSION pg_stat_statements IS 'track execution statistics of all SQL statements executed';

COMMIT;
