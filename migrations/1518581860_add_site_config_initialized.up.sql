ALTER TABLE site_config ADD COLUMN initialized boolean NOT NULL DEFAULT false;
UPDATE site_config SET initialized = EXISTS(SELECT * FROM USER);