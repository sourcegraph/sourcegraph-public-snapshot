-- Rename repo.uri -> repo.name.
ALTER TABLE repo RENAME COLUMN uri TO name;
ALTER INDEX repo_uri_unique RENAME TO repo_name_unique;
ALTER INDEX repo_uri_trgm RENAME TO repo_name_trgm;
ALTER TABLE repo RENAME CONSTRAINT check_uri_nonempty TO check_name_nonempty;

-- BACKCOMPAT: Add back a repo.uri column whose value is the same as that of repo.name. This value
-- is only set here in this migration or when a repo is updated; it is not updated whenever the name
-- column is updated. This is to avoid complexity and is sufficient because the
-- migration/backcompat-required period is usually on the order of minutes, during which time it is
-- preferable for the new version to be responsible for repository names.
--
-- This repo.uri column will be dropped one release cycle later.
ALTER TABLE repo ADD COLUMN uri citext;
UPDATE repo SET uri=name;
ALTER TABLE repo ALTER COLUMN uri SET NOT NULL;
-- BACKCOMPAT: For the same backcompat process, also set the name column's value to that of uri, in
-- case the old version executes a query that only sets the uri column. (And vice versa.)
CREATE OR REPLACE FUNCTION set_repo_name() RETURNS TRIGGER AS $$
begin
if NEW.name is null then
NEW.name := NEW.uri;
end if;
if NEW.uri is null then
NEW.uri := NEW.name;
end if;
return NEW;
end;
$$ LANGUAGE plpgsql;
CREATE TRIGGER trig_set_repo_name BEFORE INSERT ON repo FOR EACH ROW
  EXECUTE PROCEDURE set_repo_name();

-- Rename phabricator_repos.uri -> phabricator_repos.repo_name.
ALTER TABLE phabricator_repos RENAME COLUMN uri TO repo_name;
ALTER INDEX phabricator_repos_uri_key RENAME TO phabricator_repos_repo_name_key;
