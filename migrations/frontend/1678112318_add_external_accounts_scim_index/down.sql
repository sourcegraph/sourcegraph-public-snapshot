-- Undo the changes made in the up migration
DROP INDEX IF EXISTS user_external_accounts_user_id_scim_service_type;
