-- Undo the changes made in the up migration

ALTER TABLE IF EXISTS repo DROP COLUMN IF EXISTS topics;
DROP FUNCTION IF EXISTS get_topics;
