-- make the credential column nullable, because we may have credentials that are associated with a github_app instead
ALTER TABLE user_credentials
    ALTER COLUMN credential DROP NOT NULL;

ALTER TABLE user_credentials ADD COLUMN github_app_id INT NULL REFERENCES github_apps(id);
ALTER TABLE user_credentials
    ADD CONSTRAINT check_credential_and_github_app_id
        CHECK ((credential IS NOT NULL) OR (github_app_id IS NOT NULL));
