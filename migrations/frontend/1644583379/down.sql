-- Undo the changes made in the up migration
UPDATE org_invitations SET expires_at = NULL WHERE expires_at IS NOT NULL;
