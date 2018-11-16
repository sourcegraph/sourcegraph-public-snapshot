CREATE FUNCTION is_valid_json(v text) RETURNS boolean AS $$
BEGIN
  RETURN (v::json is not null);
    EXCEPTION
     WHEN others THEN RETURN false;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- It has always been impossible for manifests to contain invalid JSON, except in very early
-- versions of the extension registry (that were never shipped in a release). The Sourcegraph.com
-- database has 1 invalid manifest from this period that can be deleted. Because no other data could
-- be affected, this DELETE statement is safe.
DELETE FROM registry_extension_releases WHERE NOT is_valid_json(manifest);

ALTER TABLE registry_extension_releases ALTER COLUMN manifest TYPE jsonb USING manifest::jsonb;

DROP FUNCTION is_valid_json(text);
