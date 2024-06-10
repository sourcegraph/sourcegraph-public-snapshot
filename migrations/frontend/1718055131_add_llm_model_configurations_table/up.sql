CREATE TABLE IF NOT EXISTS model_configurations (
    created_at timestamp PRIMARY KEY NOT NULL,
    created_by INTEGER NULL REFERENCES users (id) ON DELETE SET NULL,

    base_configuration_json text,

    -- Admin supplied configuration.
    redacted_configuration_patch_json text NOT NULL,
    encrypted_configuration_patch_json text NOT NULL,
    encryption_key_id text DEFAULT ''::text NOT NULL,

    -- General configuration knobs.
    flags bigint NOT NULL
);
