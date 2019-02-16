ALTER TABLE repo ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}'
CHECK (jsonb_typeof(metadata) = 'object');
