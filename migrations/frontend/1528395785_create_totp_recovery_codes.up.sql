BEGIN;

CREATE TABLE IF NOT EXISTS totp_recovery_codes (
    id                   BIGSERIAL   PRIMARY KEY,
    user_id              INT         UNIQUE NOT NULL,
    hashed_recovery_code VARCHAR(64) NOT NULL,
    created_at           TIMESTAMP   WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP   WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT totp_recovery_codes_user_id_fk
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);

CREATE INDEX totp_recovery_codes_user_id ON totp_recovery_codes USING btree (user_id);

COMMENT ON TABLE  totp_recovery_codes                      IS 'Contains hashed TOTP recovery codes.';
COMMENT ON COLUMN totp_recovery_codes.user_id              IS 'The ID of the user that owns the recovery code.';
COMMENT ON COLUMN totp_recovery_codes.hashed_recovery_code IS 'The associated TOTP recovery code, hashed.';

COMMIT;
