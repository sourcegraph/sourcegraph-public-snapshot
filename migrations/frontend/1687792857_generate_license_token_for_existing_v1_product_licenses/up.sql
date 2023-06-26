UPDATE product_licenses
SET license_check_token = sha256(sha256(license_key :: bytea))
WHERE
    -- only update v1 licenses
    license_version = 1
    -- only update records that don't already have a license check token
    AND license_check_token IS NULL
    -- only update non-expired
    AND license_expires_at > now()
    -- only update non-revoked
    AND revoked_at IS NULL;
