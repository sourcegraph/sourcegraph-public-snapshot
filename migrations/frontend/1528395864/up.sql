BEGIN;

CREATE TABLE IF NOT EXISTS temporary_settings (
    id serial NOT NULL PRIMARY KEY,
    user_id integer NOT NULL UNIQUE,
    contents jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

COMMENT ON TABLE temporary_settings IS 'Stores per-user temporary settings used in the UI, for example, which modals have been dimissed or what theme is preferred.';
COMMENT ON COLUMN temporary_settings.user_id IS 'The ID of the user the settings will be saved for.';
COMMENT ON COLUMN temporary_settings.contents IS 'JSON-encoded temporary settings.';

COMMIT;
