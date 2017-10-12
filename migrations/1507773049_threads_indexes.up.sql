BEGIN;

CREATE INDEX threads_org_repo_id_file_branch_idx on threads (org_repo_id, file, branch);
CREATE INDEX threads_org_repo_id_branch_idx on threads (org_repo_id, branch);

COMMIT;
