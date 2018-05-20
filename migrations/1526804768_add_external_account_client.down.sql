DROP INDEX user_external_accounts_account;
UPDATE user_external_accounts SET service_id=concat(service_id, ':', client_id);
ALTER TABLE user_external_accounts DROP COLUMN client_id;
CREATE UNIQUE INDEX user_external_accounts_account ON user_external_accounts(service_type, service_id, account_id) WHERE deleted_at IS NULL;
