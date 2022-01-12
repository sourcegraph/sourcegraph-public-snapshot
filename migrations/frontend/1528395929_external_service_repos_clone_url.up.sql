-- +++
-- parent: 1528395928
-- +++

BEGIN;

CREATE INDEX IF NOT EXISTS external_service_repos_clone_url_idx ON external_service_repos (clone_url);

COMMIT;
