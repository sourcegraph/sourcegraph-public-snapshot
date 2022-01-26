-- +++
-- parent: 1528395961
-- +++

BEGIN;

ALTER TABLE IF EXISTS gitserver_repos DROP COLUMN IF EXISTS last_external_service;

COMMIT;
