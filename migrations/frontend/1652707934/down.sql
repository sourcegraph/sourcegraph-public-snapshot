-- Undo the changes made in the up migration

ALTER TABLE
    codeintel_lockfile_references
DROP
    COLUMN IF EXISTS last_check_at;
