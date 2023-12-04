DROP INDEX IF EXISTS product_licenses_license_check_token_idx;

ALTER TABLE IF EXISTS product_licenses
    DROP COLUMN IF EXISTS site_id,
    DROP COLUMN IF EXISTS license_check_token,
    DROP COLUMN IF EXISTS revoked_at,
    DROP COLUMN IF EXISTS salesforce_sub_id,
    DROP COLUMN IF EXISTS salesforce_opp_id;