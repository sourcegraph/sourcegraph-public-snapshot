BEGIN;

ALTER TABLE repo ADD COLUMN IF NOT EXISTS private BOOLEAN NOT NULL DEFAULT FALSE;

-- extract private from metadata. If we fail to extract it assume it is
-- private to avoid leaking.

DO $$
DECLARE
    t_cursor CURSOR FOR
        SELECT external_service_type, metadata FROM repo;
    t_row repo%rowtype;
    val boolean;
BEGIN
    FOR t_row IN t_cursor LOOP
        val = FALSE;
        IF t_row.external_service_type = 'github' THEN
            val = COALESCE((t_row.metadata->>'IsPrivate')::boolean, true);
        ELSIF t_row.external_service_type = 'gitlab' THEN
            val = t_row.metadata->>'visibility' <> 'public';
        ELSIF t_row.external_service_type = 'bitbucketServer' THEN
            val = not COALESCE((t_row.metadata->>'public')::boolean, false);
        ELSIF t_row.external_service_type = 'bitbucketCloud' THEN
            val = COALESCE((t_row.metadata->>'is_private')::boolean, true);
        END IF;

        UPDATE repo SET private = val WHERE CURRENT OF t_cursor;
    END LOOP;
END$$;

COMMIT;
