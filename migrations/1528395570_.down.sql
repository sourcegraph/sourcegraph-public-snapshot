ALTER TABLE repo ADD CONSTRAINT check_external CHECK (
	external_id IS NULL
  AND external_service_type IS NULL
  AND external_service_id IS NULL
  OR external_id IS NOT NULL
  AND external_service_type IS NOT NULL
  AND external_service_id IS NOT NULL
)
