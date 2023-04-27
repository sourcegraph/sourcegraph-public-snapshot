-- Add toggle
ALTER TABLE product_licenses
ADD COLUMN IF NOT EXISTS access_token_enabled BOOLEAN NOT NULL DEFAULT false;

-- In-band migration to enable usage as access tokens for existing, active license keys
UPDATE product_licenses
SET access_token_enabled = true
WHERE license_expires_at > NOW();
