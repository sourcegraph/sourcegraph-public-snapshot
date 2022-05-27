DROP INDEX IF EXISTS changesets_detached_at;

ALTER TABLE changesets DROP COLUMN detached_at;
