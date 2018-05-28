ALTER TABLE settings ADD CONSTRAINT has_subject CHECK (org_id IS NOT NULL OR user_id IS NOT NULL);
