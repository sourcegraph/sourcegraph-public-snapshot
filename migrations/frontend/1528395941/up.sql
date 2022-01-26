-- +++
-- parent: 1528395940
-- +++

BEGIN;

DELETE FROM cm_monitors WHERE namespace_user_id IS NULL;
COMMENT ON COLUMN cm_monitors.namespace_org_id IS 'DEPRECATED: code monitors cannot be owned by an org';

ALTER TABLE cm_monitors 
	ALTER COLUMN namespace_user_id SET NOT NULL;

COMMIT;
