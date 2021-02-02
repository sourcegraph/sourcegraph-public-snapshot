BEGIN;

CREATE TABLE IF NOT EXISTS totp_secrets (
    id                   BIGSERIAL   PRIMARY KEY,
    user_id              INT         UNIQUE NOT NULL,
    encrypted_secret_key VARCHAR(64) NOT NULL,
    created_at           TIMESTAMP   WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP   WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT totp_secrets_user_id_fk
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);

CREATE INDEX totp_secrets_user_id ON totp_secrets USING btree (user_id);

COMMENT ON TABLE  totp_secrets                      IS 'Contains encrypted TOTP secrets.';
COMMENT ON COLUMN totp_secrets.user_id              IS 'The ID of the user that owns the secret key.';
COMMENT ON COLUMN totp_secrets.encrypted_secret_key IS 'The associated TOTP secret key, encrypted.';

COMMIT;
