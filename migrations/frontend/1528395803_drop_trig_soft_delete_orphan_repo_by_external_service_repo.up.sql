BEGIN;

-- Drop trigger as we don't want it to fire anymore.
DROP TRIGGER IF EXISTS trig_soft_delete_orphan_repo_by_external_service_repo ON external_service_repos;
DROP FUNCTION IF EXISTS soft_delete_orphan_repo_by_external_service_repos() ;

-- Rewrite the previous function as a standalone function.
-- The function will be run manually whenever we delete an external service.
CREATE FUNCTION soft_delete_orphan_repo_by_external_service_repos() RETURNS void
    AS $$
BEGIN
    -- When an external service is soft or hard-deleted,
    -- performs a clean up to soft-delete orphan repositories.
    UPDATE
        repo
    SET
        name = soft_deleted_repository_name(name),
        deleted_at = transaction_timestamp()
    WHERE
      deleted_at IS NULL
      AND NOT EXISTS (
        SELECT FROM external_service_repos WHERE repo_id = repo.id
      );
END;
$$ LANGUAGE plpgsql;

COMMIT;
