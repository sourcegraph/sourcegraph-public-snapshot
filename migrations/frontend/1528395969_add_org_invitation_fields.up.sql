-- +++
-- parent: 1528395968
-- +++

BEGIN;

ALTER TABLE IF EXISTS org_invitations 
  ADD COLUMN IF NOT EXISTS recipient_email CITEXT,
  ADD COLUMN IF NOT EXISTS expires_at timestamp with time zone;

COMMIT;
