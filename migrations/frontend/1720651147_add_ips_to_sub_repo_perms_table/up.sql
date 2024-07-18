-- Add the new 'ips' column to the sub_repo_permissions table
ALTER TABLE IF EXISTS ONLY sub_repo_permissions
    ADD COLUMN IF NOT EXISTS ips text[];

-- Remove the check constraint to ensure ips is either NULL or has the same length as paths
ALTER TABLE IF EXISTS ONLY sub_repo_permissions
    DROP CONSTRAINT IF EXISTS ips_paths_length_check;

-- Add a check constraint to ensure ips is either NULL or has the same length as paths
ALTER TABLE IF EXISTS ONLY sub_repo_permissions
    ADD CONSTRAINT ips_paths_length_check
        CHECK (
            ips IS NULL
                OR (
                    array_length(ips, 1) = array_length(paths, 1)
                    AND NOT '' = ANY(ips) -- Don't allow empty strings
                )
            );

-- Add a comment explaining the new column
COMMENT ON COLUMN sub_repo_permissions.ips IS 'IP addresses corresponding to each path. IP in slot 0 in the array corresponds to path the in slot 0 of the path array, etc. NULL if not yet migrated, empty array for no IP restrictions.';
