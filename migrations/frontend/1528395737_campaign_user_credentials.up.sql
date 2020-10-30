BEGIN;

CREATE TABLE IF NOT EXISTS campaign_user_credentials (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    external_service_type TEXT NOT NULL,
    external_service_id TEXT NOT NULL,
    credential TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- deleted_at is intentionally left out: given the secret nature of tokens,
    -- users will likely want to be certain that tokens are deleted when
    -- removed, even with encryption in place.

    -- Set up the foreign key.
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE DEFERRABLE
);

COMMIT;
