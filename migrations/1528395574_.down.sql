BEGIN;

ALTER TABLE repo DROP CONSTRAINT IF EXISTS repo_name_unique;

CREATE UNIQUE INDEX repo_name_unique ON repo(name);

COMMIT;
