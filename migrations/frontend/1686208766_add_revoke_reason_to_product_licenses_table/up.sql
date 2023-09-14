ALTER TABLE IF EXISTS product_licenses 
    ADD COLUMN IF NOT EXISTS revoke_reason TEXT;
