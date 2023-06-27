UPDATE product_licenses
SET license_check_token = sha256(sha256(license_key :: bytea))
WHERE
    -- update v1 licenses where token is NULL
    ((
        -- only update v1 licenses
        license_version = 1 -- only update records that don't already have a license check token
        AND license_check_token IS NULL
    )
    -- update v2 licenses where token does not match expectation
    OR (
        license_version = 2
        AND license_check_token != sha256(sha256(license_key :: bytea))
    ))
    -- only update non-expired
    AND license_expires_at > now()
    -- only update non-revoked
    AND revoked_at IS NULL;
