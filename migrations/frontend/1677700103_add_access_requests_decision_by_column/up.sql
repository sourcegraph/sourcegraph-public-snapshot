ALTER TABLE IF EXISTS access_requests
ADD COLUMN IF NOT EXISTS decision_by INT NULL REFERENCES users(id) ON DELETE
SET NULL;
