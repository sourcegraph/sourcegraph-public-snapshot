BEGIN;

-- We don't remove the out of band migration when moving down.

ALTER TABLE
    external_services
DROP COLUMN IF EXISTS
    has_webhooks;

COMMIT;
