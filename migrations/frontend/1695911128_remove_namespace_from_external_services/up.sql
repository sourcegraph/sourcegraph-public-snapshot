DROP TRIGGER IF EXISTS trig_soft_delete_user_reference_on_external_service ON users;
DROP FUNCTION IF EXISTS soft_delete_user_reference_on_external_service;
DROP INDEX IF EXISTS external_services_unique_kind_org_id;
DROP INDEX IF EXISTS external_services_unique_kind_user_id;
ALTER TABLE external_service_repos DROP COLUMN IF EXISTS user_id;
ALTER TABLE external_service_repos DROP COLUMN IF EXISTS org_id;
ALTER TABLE external_services DROP COLUMN IF EXISTS namespace_user_id;
ALTER TABLE external_services DROP COLUMN IF EXISTS namespace_org_id;
