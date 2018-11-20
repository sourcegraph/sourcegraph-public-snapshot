BEGIN;
ALTER TABLE site_config RENAME TO global_state;
ALTER INDEX site_config_pkey RENAME TO global_state_pkey;
CREATE VIEW site_config AS SELECT * FROM global_state;
END;
