ALTER TABLE global_state
ADD
    COLUMN IF NOT EXISTS is_license_valid boolean NULL;
