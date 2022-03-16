ALTER TABLE IF EXISTS gitserver_repos ADD COLUMN IF NOT EXISTS last_external_service bigint;
