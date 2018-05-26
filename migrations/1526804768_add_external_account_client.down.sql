DROP INDEX user_external_accounts_account;
CREATE UNIQUE INDEX user_external_accounts_account ON user_external_accounts(service_type, service_id, account_id) WHERE deleted_at IS NULL;

ALTER TABLE user_external_accounts DROP COLUMN client_id;
