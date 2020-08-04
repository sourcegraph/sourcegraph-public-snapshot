-- Note: CREATE INDEX CONCURRENTLY cannot run inside a transaction block
CREATE INDEX CONCURRENTLY IF NOT EXISTS repo_name_idx ON public.repo USING btree (lower(name::text) COLLATE pg_catalog."C");
