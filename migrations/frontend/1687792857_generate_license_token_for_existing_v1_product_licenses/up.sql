UPDATE product_licenses
SET license_check_token = sha256(license_key :: bytea) 
