ALTER TABLE site_config DROP COLUMN id;
ALTER TABLE site_config RENAME COLUMN app_id TO site_id;
ALTER TABLE site_config ADD PRIMARY KEY (site_id);
ALTER TABLE site_config RENAME COLUMN last_updated TO updated_at;
ALTER TABLE site_config ALTER COLUMN updated_at TYPE timestamp with time zone
	USING to_timestamp(updated_at, 'YYYY-MM-DD HH24:MI:SS');
