BEGIN;

ALTER TABLE repo ADD COLUMN IF NOT EXISTS private BOOLEAN NOT NULL DEFAULT FALSE;

DO $$
DECLARE
    t_cursor CURSOR FOR
        SELECT id, external_service_type, metadata FROM repo;
    t_row repo%rowtype;
    val boolean;
BEGIN
    FOR t_row IN t_cursor LOOP
        val = FALSE;
        IF t_row.external_service_type = 'github' THEN
            val = t_row.metadata ->> 'IsPrivate';
        ELSIF t_row.external_service_type = 'gitlab' THEN
            IF t_row.metadata ->> 'visibility' = 'private' THEN
                val = TRUE;
            END IF;
        ELSIF t_row.external_service_type = 'bitbucketServer' THEN
            val = NOT CAST(t_row.metadata ->> 'public' AS BOOLEAN);
        END IF;

        UPDATE repo SET private = val WHERE current of t_cursor;
    END LOOP;
END$$;

COMMIT;
