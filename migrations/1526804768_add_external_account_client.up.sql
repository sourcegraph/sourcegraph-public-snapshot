ALTER TABLE user_external_accounts ADD COLUMN client_id text;
UPDATE user_external_accounts SET client_id='';
ALTER TABLE user_external_accounts ALTER client_id SET NOT NULL;
DROP INDEX user_external_accounts_account;
CREATE UNIQUE INDEX user_external_accounts_account ON user_external_accounts(service_type, service_id, client_id, account_id) WHERE deleted_at IS NULL;
