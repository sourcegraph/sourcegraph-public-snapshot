DO $$
BEGIN
    ALTER TABLE public.product_licenses
        RENAME COLUMN license_key_sha256  TO license_check_token;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column license_key_sha256 does not exist in table product_licenses';
END $$;

-- Forcibly set all license_check_token to the old format of double-hashing
UPDATE product_licenses
SET license_check_token = sha256(sha256(license_key :: bytea))
WHERE license_check_token IS NOT NULL;

-- Recreate the old index on the renamed column
DROP INDEX IF EXISTS product_licenses_license_key_sha256_idx;
CREATE UNIQUE INDEX IF NOT EXISTS product_licenses_license_check_token_idx ON product_licenses(license_check_token);
