-- Undo the changes made in the up migration
ALTER TABLE IF EXISTS lsif_references
    ALTER COLUMN filter SET NOT NULL;
