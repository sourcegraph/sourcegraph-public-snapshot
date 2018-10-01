ALTER TABLE site_config DROP CONSTRAINT site_config_pkey;
ALTER TABLE site_config ADD COLUMN id serial;
ALTER TABLE site_config ADD PRIMARY KEY (id);
ALTER TABLE site_config RENAME COLUMN site_id TO app_id;
ALTER TABLE site_config RENAME COLUMN updated_at TO last_updated;
ALTER TABLE site_config ALTER COLUMN last_updated TYPE text
	USING to_char(last_updated, 'YYYY-MM-DD HH24:MI:SS');
