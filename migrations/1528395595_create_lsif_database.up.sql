-- Enable extension that allow us to perform SQL queries within another
-- connection context. This is necessary so that we can run migrations
-- on the <sg>_lsif database.
-- Note: `$$$PGPASSWORD$$$` is replaced by the frontend with the real
--        password for the currently authed user before migrations run.

CREATE EXTENSION IF NOT EXISTS dblink;

-- This function runs the given SQL query as the current user in the
-- database named {current_db}{suffix}. If suffix is empty, then it
-- is run in the current database OUTSIDE of any existing transaction.

CREATE OR REPLACE FUNCTION remote_exec(suffix text, query text) RETURNS void AS $$
BEGIN
    PERFORM dblink_exec('dbname=' || current_database() || suffix || ' user=' || current_user || ' password=$$$PGPASSWORD$$$', query);
END;
$$
LANGUAGE plpgsql;

-- Create the LSIF database. We can't run this particular query within a
-- transaction which, for some reason, postgres and/or golang-migrate thinks
-- we're in. See https://github.com/golang-migrate/migrate/issues/284 for
-- additional details. To issue this query in a non-transactional context,
-- we use dblink.

CREATE OR REPLACE FUNCTION create_lsif_db() RETURNS void AS $$
BEGIN
    -- Ensure db doesn't already exist before trying to create it.
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_database WHERE datname = current_database() || '_lsif') THEN
        PERFORM remote_exec('', 'CREATE DATABASE "' || current_database() || '_lsif" OWNER DEFAULT TEMPLATE template0 ENCODING ''UTF8''');
    END IF;
END;
$$ LANGUAGE plpgsql;

SELECT create_lsif_db();
DROP FUNCTION create_lsif_db();

-- Ensure dblink extension is enabled on the other side as well. This is
-- necessary as the LSIF processes check the migration state on startup,
-- which is in the sourcegraph database.

SELECT remote_exec('_lsif', 'CREATE EXTENSION IF NOT EXISTS dblink;');
