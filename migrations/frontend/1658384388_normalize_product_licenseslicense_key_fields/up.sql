ALTER TABLE product_licenses ADD COLUMN IF NOT EXISTS license_version INT;
ALTER TABLE product_licenses ADD COLUMN IF NOT EXISTS license_tags TEXT[];
ALTER TABLE product_licenses ADD COLUMN IF NOT EXISTS license_user_count INT;
ALTER TABLE product_licenses ADD COLUMN IF NOT EXISTS license_expires_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE product_subscriptions ADD COLUMN IF NOT EXISTS account_number TEXT;
