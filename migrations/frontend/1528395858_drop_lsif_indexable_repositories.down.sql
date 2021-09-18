
BEGIN;

CREATE TABLE lsif_indexable_repositories (
    id SERIAL PRIMARY KEY NOT NULL,
    repository_id integer NOT NULL,
    search_count integer DEFAULT 0 NOT NULL,
    precise_count integer DEFAULT 0 NOT NULL,
    last_index_enqueued_at timestamp with time zone,
    last_updated_at timestamp with time zone DEFAULT now() NOT NULL,
    enabled boolean
);

CREATE UNIQUE INDEX lsif_indexable_repositories_repository_id_key ON lsif_indexable_repositories (repository_id);

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
