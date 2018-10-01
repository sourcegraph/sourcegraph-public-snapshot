ALTER TABLE deployment_configuration RENAME TO site_config;
ALTER TABLE site_config RENAME CONSTRAINT deployment_configuration_pkey TO site_config_pkey;
