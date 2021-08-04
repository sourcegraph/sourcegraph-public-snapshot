BEGIN;

DO $$
BEGIN
    REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM sg_service;
    REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM sg_service;
    REVOKE USAGE ON SCHEMA public FROM sg_service;
EXCEPTION WHEN undefined_object THEN
    -- Treat this like an "IF EXISTS": if for some reason the role isn't
    -- present, proceed to subsequent migration steps.
END;
$$;

COMMIT;
