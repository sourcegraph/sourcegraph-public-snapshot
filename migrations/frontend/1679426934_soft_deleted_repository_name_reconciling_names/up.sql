CREATE OR REPLACE FUNCTION soft_deleted_repository_name(name text) RETURNS text
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF name LIKE 'DELETED-%' THEN
        RETURN name;
    ELSE
        RETURN 'DELETED-' || extract(epoch from transaction_timestamp()) || '-' || name;
    END IF;
END;
$$;
