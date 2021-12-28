ALTER TABLE search_contexts ADD COLUMN repo_query TEXT;

CREATE INDEX search_contexts_repo_query_idx ON search_contexts USING BTREE (repo_query);
