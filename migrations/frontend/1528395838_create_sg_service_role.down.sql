BEGIN;

DO $$
BEGIN
    DROP OWNED BY sg_service;
    DROP ROLE IF EXISTS sg_service;
EXCEPTION WHEN dependent_objects_still_exist THEN
    -- Roles are cluster-wide, which makes them visible to both real and test
    -- code. Since tests run in the same cluster as local development code, it
    -- may not be possible to roll the change back.
END;
$$;

COMMIT;
