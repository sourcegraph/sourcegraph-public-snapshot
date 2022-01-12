-- +++
-- parent: 1528395861
-- +++

BEGIN;

DO $$
BEGIN
    DROP ROLE IF EXISTS sg_service;
EXCEPTION WHEN dependent_objects_still_exist THEN
    -- Roles are visible across databases within a server, and we use templated
    -- databases for test parallelism, so it's possible in some cases for the
    -- tests to hit a case where the role can't be dropped because one of the
    -- test databases still has objects that depend on it.
END;
$$;

COMMIT;
