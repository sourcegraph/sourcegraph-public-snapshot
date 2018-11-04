DROP TRIGGER IF EXISTS trig_set_repo_name ON repo;
DROP FUNCTION IF EXISTS set_repo_name();
ALTER TABLE repo DROP COLUMN uri;

ALTER TABLE repo RENAME COLUMN name TO uri;
ALTER INDEX repo_name_unique RENAME TO repo_uri_unique;
ALTER INDEX repo_name_trgm RENAME TO repo_uri_trgm;
ALTER TABLE repo RENAME CONSTRAINT check_name_nonempty TO check_uri_nonempty;

ALTER TABLE phabricator_repos RENAME COLUMN repo_name TO uri;
ALTER INDEX phabricator_repos_repo_name_key RENAME TO phabricator_repos_uri_key;
