-- A prior migration (1528395595_create_lsif_database.up.sql) created this
-- function and was removed (see that migration file for details), so here we
-- clean up by removing the function. We do not delete the `${PGDATABASE}_lsif`
-- DB here since that would require using dblink. This only ran in dev
-- deployments and sourcegraph.com, where we can just drop it manually.

DROP FUNCTION IF EXISTS remote_exec(text, text);
