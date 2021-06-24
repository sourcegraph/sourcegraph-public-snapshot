BEGIN;

DO $$
BEGIN
    REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM sg_service;
    REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM sg_service;
    REVOKE USAGE ON SCHEMA public FROM sg_service;
    DROP ROLE IF EXISTS sg_service;
EXCEPTION WHEN dependent_objects_still_exist THEN
    -- Roles are cluster-wide, which makes them visible to both real and test
    -- code. Since tests run in the same cluster as local development code, it
    -- may not be possible to roll the change back.
END;
$$;

COMMIT;
