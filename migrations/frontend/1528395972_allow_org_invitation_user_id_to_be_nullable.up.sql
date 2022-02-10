-- +++
-- parent: 1528395971
-- +++

BEGIN;

ALTER TABLE IF EXISTS org_invitations ALTER COLUMN recipient_user_id DROP NOT NULL;
ALTER TABLE IF EXISTS org_invitations
  ADD CONSTRAINT either_user_id_or_email_defined CHECK ((recipient_user_id IS NULL) != (recipient_email IS NULL));

COMMIT;
