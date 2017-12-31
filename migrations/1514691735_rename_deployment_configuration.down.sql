ALTER TABLE site_config RENAME TO deployment_configuration;
ALTER TABLE site_config RENAME CONSTRAINT site_config_pkey TO deployment_configuration_pkey;
