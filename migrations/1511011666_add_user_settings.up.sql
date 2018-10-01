ALTER TABLE settings ADD COLUMN user_id integer REFERENCES users (id) ON DELETE RESTRICT;

-- Relax has_subject constraint to also allow user settings.
ALTER TABLE settings DROP CONSTRAINT has_subject;
ALTER TABLE settings ADD CONSTRAINT has_subject CHECK (org_id IS NOT NULL OR user_id IS NOT NULL);
