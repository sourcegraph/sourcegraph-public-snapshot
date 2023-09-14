-- Add toggle
ALTER TABLE product_licenses
ADD COLUMN IF NOT EXISTS access_token_enabled BOOLEAN NOT NULL DEFAULT false;

-- Documentation!
COMMENT ON COLUMN product_licenses.access_token_enabled
IS 'Whether this license key can be used as an access token to authenticate API requests';

-- In-band migration to enable usage as access tokens for existing, active license keys
UPDATE product_licenses
SET access_token_enabled = true
WHERE license_expires_at > NOW();
