BEGIN;

DROP INDEX threads_org_repo_id_file_branch_idx;
DROP INDEX threads_org_repo_id_branch_idx;

COMMIT;
