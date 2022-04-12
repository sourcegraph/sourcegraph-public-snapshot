ALTER TABLE IF EXISTS org_invitations
  DROP COLUMN IF EXISTS recipient_email,
  DROP COLUMN IF EXISTS expires_at;
