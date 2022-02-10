BEGIN;

-- Undo the changes made in the up migration
ALTER TABLE IF EXISTS org_invitations ALTER COLUMN recipient_user_id SET NOT NULL;
ALTER TABLE IF EXISTS org_invitations DROP CONSTRAINT either_user_id_or_email_defined;

COMMIT;
