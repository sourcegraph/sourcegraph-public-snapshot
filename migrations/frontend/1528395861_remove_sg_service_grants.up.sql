
BEGIN;

DO $$
BEGIN
    REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM sg_service;
    REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM sg_service;
    REVOKE USAGE ON SCHEMA public FROM sg_service;
EXCEPTION WHEN undefined_object THEN
    -- Roles are visible across databases within a server, and we use templated
    -- databases for test parallelism, so it's possible in some cases for the
    -- tests to hit a case where the role can't be dropped because one of the
    -- test databases still has objects that depend on it.
END;
$$;

COMMIT;
