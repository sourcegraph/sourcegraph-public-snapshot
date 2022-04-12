ALTER TABLE search_contexts ADD COLUMN IF NOT EXISTS query TEXT;

CREATE INDEX IF NOT EXISTS search_contexts_query_idx ON search_contexts USING BTREE (query);
