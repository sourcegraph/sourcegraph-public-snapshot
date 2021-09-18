
BEGIN;

ALTER TABLE gitserver_repos ADD COLUMN last_fetched TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now();

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
