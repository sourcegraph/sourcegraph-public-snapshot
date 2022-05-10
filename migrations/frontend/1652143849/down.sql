-- Undo the changes made in the up migration
ALTER TABLE lsif_references ADD COLUMN IF NOT EXISTS filter bytea NOT NULL DEFAULT '{}'::bytea;
