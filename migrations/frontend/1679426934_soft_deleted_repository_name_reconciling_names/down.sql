CREATE OR REPLACE FUNCTION soft_deleted_repository_name(name text) RETURNS text
    LANGUAGE plpgsql STRICT
    AS $$
BEGIN
    RETURN 'DELETED-' || extract(epoch from transaction_timestamp()) || '-' || name;
END;
$$;
