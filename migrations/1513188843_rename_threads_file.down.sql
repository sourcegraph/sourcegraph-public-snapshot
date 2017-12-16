BEGIN;

DROP INDEX threads_org_repo_id_repo_revision_path_branch_idx;
DROP INDEX threads_org_repo_id_lines_revision_path_branch_idx;

ALTER TABLE threads ALTER COLUMN repo_revision_path DROP NOT NULL;
ALTER TABLE threads DROP COLUMN lines_revision_path;
ALTER TABLE threads RENAME COLUMN repo_revision_path TO "file";

CREATE INDEX threads_local_repo_id_file_idx ON threads(org_repo_id, "file");
CREATE INDEX threads_org_repo_id_file_branch_idx ON threads(org_repo_id, "file", branch);

COMMIT;
