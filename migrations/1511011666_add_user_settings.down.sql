ALTER TABLE settings DROP CONSTRAINT has_subject;
ALTER TABLE settings ADD CONSTRAINT has_subject CHECK (org_id IS NOT NULL);
ALTER TABLE settings DROP COLUMN user_id;
