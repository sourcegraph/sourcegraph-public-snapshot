DROP INDEX IF EXISTS changesets_detached_at;

ALTER TABLE changesets DROP COLUMN IF EXISTS detached_at;
