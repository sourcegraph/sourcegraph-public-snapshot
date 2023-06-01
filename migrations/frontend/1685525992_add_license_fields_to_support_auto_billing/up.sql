ALTER TABLE IF EXISTS product_licenses
    ADD COLUMN IF NOT EXISTS site_id UUID NULL,
    ADD COLUMN IF NOT EXISTS license_check_token bytea NULL,
    ADD COLUMN IF NOT EXISTS revoked_at timestamptz NULL,
    ADD COLUMN IF NOT EXISTS salesforce_sub_id text NULL,
    ADD COLUMN IF NOT EXISTS salesforce_opp_id text NULL;

CREATE UNIQUE INDEX IF NOT EXISTS product_licenses_license_check_token_idx ON product_licenses(license_check_token);
