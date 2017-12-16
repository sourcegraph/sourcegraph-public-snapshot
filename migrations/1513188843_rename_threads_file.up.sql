BEGIN;

DROP INDEX threads_local_repo_id_file_idx;
DROP INDEX threads_org_repo_id_file_branch_idx;

ALTER TABLE threads RENAME COLUMN "file" TO repo_revision_path;
ALTER TABLE threads ADD COLUMN lines_revision_path TEXT;

UPDATE threads SET lines_revision_path = repo_revision_path;

ALTER TABLE threads ALTER COLUMN repo_revision_path SET NOT NULL;
ALTER TABLE threads ALTER COLUMN lines_revision_path SET NOT NULL;

CREATE INDEX threads_org_repo_id_repo_revision_path_branch_idx ON threads(org_repo_id, repo_revision_path, branch);
CREATE INDEX threads_org_repo_id_lines_revision_path_branch_idx ON threads(org_repo_id, lines_revision_path, branch);

COMMIT;