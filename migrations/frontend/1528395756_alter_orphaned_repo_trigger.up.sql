BEGIN;

DROP FUNCTION IF EXISTS soft_delete_orphan_repo_by_external_service_repos() CASCADE;
DROP TRIGGER IF EXISTS trig_soft_delete_orphan_repo_by_external_service_repo ON external_service_repos;

CREATE FUNCTION soft_delete_orphan_repo_by_external_service_repos() RETURNS trigger
    LANGUAGE plpgsql
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

    RETURN NULL;
END;
$$;

CREATE TRIGGER trig_soft_delete_orphan_repo_by_external_service_repo
    AFTER DELETE ON external_service_repos
    FOR EACH STATEMENT EXECUTE PROCEDURE soft_delete_orphan_repo_by_external_service_repos();

COMMIT;
