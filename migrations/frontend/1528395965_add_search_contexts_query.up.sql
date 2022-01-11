ALTER TABLE search_contexts ADD COLUMN query TEXT;

CREATE INDEX search_contexts_query_idx ON search_contexts USING BTREE (query);
