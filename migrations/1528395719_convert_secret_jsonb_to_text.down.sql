BEGIN;

-- ALTER TABLE user_external_accounts ALTER COLUMN auth_data TYPE JSONB USING to_jsonb(auth_data);
-- ALTER TABLE user_external_accounts ALTER COLUMN account_data TYPE JSONB USING to_jsonb(account_data);
-- ALTER TABLE repo ALTER COLUMN metadata TYPE JSONB USING to_jsonb(metadata);
-- ALTER TABLE repo ALTER COLUMN metadata SET DEFAULT '{}'::JSONB;
-- CREATE INDEX repo_metadata_gin_idx ON repo USING gin (metadata);
-- ALTER TABLE repo ADD CONSTRAINT repo_metadata_check CHECK (jsonb_typeof(metadata) = 'object'::text);

COMMIT;
