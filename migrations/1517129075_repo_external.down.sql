ALTER TABLE repo DROP CONSTRAINT check_external;
ALTER TABLE repo RENAME COLUMN external_id TO origin_repo_id;
ALTER TABLE repo RENAME COLUMN external_service_type TO origin_service;
ALTER TABLE repo ALTER COLUMN origin_service TYPE integer USING CASE WHEN origin_service='github' THEN 0 ELSE null END;
ALTER TABLE repo RENAME COLUMN external_service_id TO origin_api_base_url;
