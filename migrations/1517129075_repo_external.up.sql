ALTER TABLE repo RENAME COLUMN origin_repo_id TO external_id;
ALTER TABLE repo RENAME COLUMN origin_service TO external_service_type;
ALTER TABLE repo ALTER COLUMN external_service_type TYPE text USING CASE WHEN external_service_type=0 THEN 'github' ELSE null END;
ALTER TABLE repo RENAME COLUMN origin_api_base_url TO external_service_id;
ALTER TABLE repo ADD CONSTRAINT check_external CHECK ((external_id IS NULL AND external_service_type IS NULL AND external_service_id IS NULL) OR (external_id IS NOT NULL AND external_service_type IS NOT NULL AND external_service_id IS NOT NULL));
COMMIT;