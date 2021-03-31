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

CREATE TRIGGER trig_delete_external_service_ref_on_external_service_repos
    AFTER UPDATE OF deleted_at ON external_services FOR EACH ROW EXECUTE PROCEDURE delete_external_service_ref_on_external_service_repos();

COMMIT;
