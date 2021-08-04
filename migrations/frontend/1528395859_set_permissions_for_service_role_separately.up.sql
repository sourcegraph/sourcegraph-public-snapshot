BEGIN;

DO $$
BEGIN
    GRANT USAGE ON SCHEMA public TO sg_service;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO sg_service;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO sg_service;
EXCEPTION WHEN undefined_object THEN
    -- Treat this like an "IF EXISTS": if for some reason the role isn't
    -- present, proceed to subsequent migration steps.
END;
$$;

COMMIT;
