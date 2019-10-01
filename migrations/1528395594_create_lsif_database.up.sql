-- dblink allows us to query other databases without opening a second connection.
-- This is used in this context to run migrations for the LSIF database.

CREATE EXTENSION IF NOT EXISTS dblink;

-- We can't run this particular query within a transaction which, for some reason,
-- postgres/golang-migrate thinks we're in. See the following issue for additional
-- details: https://github.com/golang-migrate/migrate/issues/284. To issue this
-- query in a non-transactional context, we use dblink.

SELECT dblink_exec('dbname=' || current_database() || ' user=' || current_user, '
    CREATE DATABASE sourcegraph_lsif OWNER DEFAULT TEMPLATE template0 ENCODING ''UTF8''
');
