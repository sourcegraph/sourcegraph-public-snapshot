BEGIN;

-- need to ensure we cascade delete repo. This table has a foreign key but
-- doesn't specify an action on delete.
ALTER TABLE IF EXISTS default_repos
    DROP CONSTRAINT default_repos_repo_id_fkey,
    ADD CONSTRAINT default_repos_repo_id_fkey FOREIGN KEY (repo_id) REFERENCES repo(id) ON DELETE CASCADE;

-- drop any remaining rows that haven't migrated. The only place this should
-- have any action is on Sourcegraph.com
DELETE FROM repo WHERE ((external_service_type IS NULL) OR (external_service_id IS NULL) OR (external_id IS NULL));

-- repo_external_service_unique_idx is currently a partial index, we want to
-- make it a full index. To do that we need to create a full index and only
-- then drop. If we did the other order we could accidentally introduce
-- duplicates.

CREATE UNIQUE INDEX IF NOT EXISTS repo_external_unique_idx ON repo USING btree (external_service_type, external_service_id, external_id);

DROP INDEX IF EXISTS repo_external_service_unique_idx;

COMMIT;
