BEGIN;

ALTER TABLE IF EXISTS org_invitations 
  ADD COLUMN recipient_email CITEXT,
  ADD COLUMN expires_at timestamp with time zone;

COMMIT;
