-- Add column
ALTER TABLE product_licenses
ADD COLUMN IF NOT EXISTS access_token_sha256 BYTEA;

-- Make sure all values are unqiue
ALTER TABLE product_licenses
    DROP CONSTRAINT IF EXISTS product_licenses_access_token_sha256_unique,
    ADD CONSTRAINT product_licenses_access_token_sha256_unique UNIQUE (access_token_sha256);

-- Index for lookups
CREATE INDEX IF NOT EXISTS product_licenses_access_token_sha256_idx
ON product_licenses (access_token_sha256)
WHERE access_token_sha256 IS NOT NULL;

-- In-band migration to create access_token_sha256 for existing, active license keys using the initial format
-- The token is a hash of license_key, so access_token_sha256 is a hash of that
UPDATE product_licenses
SET access_token_sha256 = digest(digest(license_key, 'sha256'), 'sha256')
WHERE
    license_expires_at > NOW()
    AND access_token_sha256 IS NULL;
