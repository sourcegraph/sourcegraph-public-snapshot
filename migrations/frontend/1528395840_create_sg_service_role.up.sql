BEGIN;

-- The "sg_service" role is one that the frontend and other services
-- will assume on startup/init. It lowers the privilege of those services
-- such that we can apply security policies to the role and let Postgres
-- manage things that previously would need to be done in app-level code.
DO $$
BEGIN
    CREATE ROLE sg_service INHERIT;
    GRANT USAGE ON SCHEMA public TO sg_service;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO sg_service;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO sg_service;
EXCEPTION 
	WHEN duplicate_object THEN
    -- Roles are cluster-wide, which makes them visible to both real and test
    -- code. The test runners may effectively execute this code multiple times,
    -- so if the role happens to exist, we just ignore it.
	WHEN unique_violation THEN
    -- Same as above.
END;
$$;

COMMIT;
