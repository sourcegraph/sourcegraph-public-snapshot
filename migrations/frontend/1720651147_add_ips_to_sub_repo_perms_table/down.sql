-- Remove the check constraint to ensure ips is either NULL or has the same length as paths
ALTER TABLE IF EXISTS ONLY sub_repo_permissions
    DROP CONSTRAINT IF EXISTS ips_paths_length_check;

-- Remove the new 'ips' column from the sub_repo_permissions table
ALTER TABLE IF EXISTS ONLY sub_repo_permissions
    DROP COLUMN IF EXISTS ips;
