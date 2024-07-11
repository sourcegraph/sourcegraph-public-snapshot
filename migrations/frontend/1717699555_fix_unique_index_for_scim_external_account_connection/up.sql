DROP INDEX IF EXISTS user_external_accounts_user_id_scim_service_type;

CREATE UNIQUE INDEX
    user_external_accounts_user_id_scim_service_type
ON
    user_external_accounts (user_id, service_type)
WHERE
    service_type = 'scim'::text
    AND deleted_at IS NULL;
