BEGIN;

ALTER TABLE IF EXISTS external_services ADD COLUMN IF NOT EXISTS config_bin bytea DEFAULT '\x00'
    NOT NULL;

ALTER TABLE IF EXISTS event_logs ADD COLUMN  IF NOT EXISTS argument_bin bytea DEFAULT '\x00'
NOT NULL;


ALTER TABLE IF EXISTS user_external_accounts ADD COLUMN IF NOT EXISTS auth_data_bin bytea DEFAULT '\x00' NOT NULL , ADD COLUMN account_data_bin bytea DEFAULT '\x00' NOT NULL;


COMMIT;
