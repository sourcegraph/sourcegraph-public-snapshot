-- Rename the column to something that reflects the desired contents
DO $$
BEGIN
    ALTER TABLE public.product_licenses
        RENAME COLUMN license_check_token TO license_key_sha256;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column license_check_token does not exist in table product_licenses';
END $$;

-- Forcibly set all license_key_sha256 to sha256 of the license key
UPDATE product_licenses
SET license_key_sha256 = sha256(license_key :: bytea)
WHERE license_key_sha256 IS NOT NULL;

-- Recreate the index on the renamed column
DROP INDEX IF EXISTS product_licenses_license_check_token_idx;
CREATE UNIQUE INDEX IF NOT EXISTS product_licenses_license_key_sha256_idx ON product_licenses(license_key_sha256);
