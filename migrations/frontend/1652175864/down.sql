ALTER TABLE
    repo_permissions
    DROP
        COLUMN IF EXISTS unrestricted;

DROP INDEX IF EXISTS repo_permissions_unrestricted_true_idx;
