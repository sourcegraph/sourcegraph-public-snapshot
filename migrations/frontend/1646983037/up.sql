BEGIN;

CREATE TABLE IF NOT EXISTS gitserver_repo_migration_cursor (
    cursor text not null
);

COMMIT;
