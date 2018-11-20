BEGIN;
DROP VIEW site_config;
ALTER TABLE global_state RENAME TO site_config;
ALTER INDEX global_state_pkey RENAME TO site_config_pkey;
END;
