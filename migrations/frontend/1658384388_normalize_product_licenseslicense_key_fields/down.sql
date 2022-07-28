ALTER TABLE product_licenses DROP COLUMN IF EXISTS license_version;
ALTER TABLE product_licenses DROP COLUMN IF EXISTS license_tags;
ALTER TABLE product_licenses DROP COLUMN IF EXISTS license_user_count;
ALTER TABLE product_licenses DROP COLUMN IF EXISTS license_expires_at;
ALTER TABLE product_subscriptions DROP COLUMN IF EXISTS account_number;
