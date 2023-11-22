ALTER TABLE github_app_installs
    DROP CONSTRAINT IF EXISTS unique_app_install;

ALTER TABLE github_app_installs ADD CONSTRAINT unique_app_install UNIQUE (app_id, installation_id);
