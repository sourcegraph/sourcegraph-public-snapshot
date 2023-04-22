ALTER TABLE product_licenses
DROP CONSTRAINT IF EXISTS product_licenses_access_token_sha256_unique;

DROP INDEX IF EXISTS access_token_sha256_idx;

ALTER TABLE product_licenses
DROP COLUMN IF EXISTS access_token_sha256;
