BEGIN;

CREATE FUNCTION delete_external_service_ref_on_external_service_repos() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        -- if an external service is soft-deleted, delete every row that references it
        IF (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL) THEN
            DELETE FROM
                external_service_repos
            WHERE
                external_service_id = OLD.id;
        END IF;

        RETURN OLD;
    END;
$$;

CREATE FUNCTION delete_repo_ref_on_external_service_repos() RETURNS trigger
    LANGUAGE plpgsql
AS $$
BEGIN
    -- if a repo is soft-deleted, delete every row that references that repo
    IF (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL) THEN
        DELETE FROM
            external_service_repos
        WHERE
            repo_id = OLD.id;
    END IF;

    RETURN OLD;
END;
$$;

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
        AND id = OLD.repo_id
        AND id NOT IN (
            SELECT DISTINCT(repo_id) FROM external_service_repos
        );

    RETURN OLD;
END;
$$;

CREATE FUNCTION soft_delete_user_reference_on_external_service() RETURNS trigger
    LANGUAGE plpgsql
AS $$
BEGIN
    -- If a user is soft-deleted, delete every row that references that user
    IF (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL) THEN
        UPDATE external_services
        SET deleted_at = NOW()
        WHERE namespace_user_id = OLD.id;
    END IF;

    RETURN OLD;
END;
$$;

CREATE TRIGGER trig_delete_external_service_ref_on_external_service_repos AFTER UPDATE OF deleted_at ON external_services FOR EACH ROW EXECUTE PROCEDURE delete_external_service_ref_on_external_service_repos();

CREATE TRIGGER trig_delete_repo_ref_on_external_service_repos AFTER UPDATE OF deleted_at ON repo FOR EACH ROW EXECUTE PROCEDURE delete_repo_ref_on_external_service_repos();

CREATE TRIGGER trig_soft_delete_orphan_repo_by_external_service_repo AFTER DELETE ON external_service_repos FOR EACH ROW EXECUTE PROCEDURE soft_delete_orphan_repo_by_external_service_repos();

CREATE TRIGGER trig_soft_delete_user_reference_on_external_service AFTER UPDATE OF deleted_at ON users FOR EACH ROW EXECUTE PROCEDURE soft_delete_user_reference_on_external_service();

COMMIT;

