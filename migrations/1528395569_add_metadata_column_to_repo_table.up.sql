ALTER TABLE repo ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}'
CHECK (jsonb_typeof(metadata) = 'object');
CREATE INDEX repo_metadata_gin_idx ON repo USING GIN (metadata);
