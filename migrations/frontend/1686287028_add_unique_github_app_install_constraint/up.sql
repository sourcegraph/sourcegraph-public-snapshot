ALTER TABLE github_app_installs
ADD CONSTRAINT unique_app_install
UNIQUE (app_id, installation_id);
