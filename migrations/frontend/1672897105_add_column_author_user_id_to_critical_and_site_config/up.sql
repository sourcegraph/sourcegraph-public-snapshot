ALTER TABLE critical_and_site_config
      ADD COLUMN IF NOT EXISTS author_user_id integer;
COMMENT ON COLUMN critical_and_site_config.author_user_id IS 'A null value indicates that this config was most likely added by code on the start-up path, for example from the SITE_CONFIG_FILE unless the config itself was added before this column existed in which case it could also have been a user.';
