-- Undo the changes made in the up migration
ALTER TABLE IF EXISTS users ALTER COLUMN invite_quota SET DEFAULT 15;
