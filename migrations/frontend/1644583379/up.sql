-- If we have invitations that are not deleted, not revoked, not responded to and with no expiry time
-- set the default expiry time to 7 days from now
UPDATE
    org_invitations
SET
    expires_at = now() + interval '7 days'
WHERE
    deleted_at IS NULL
    AND revoked_at IS NULL
    AND responded_at IS NULL
    AND expires_at IS NULL;
