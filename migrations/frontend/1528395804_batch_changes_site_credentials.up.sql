BEGIN;

CREATE TABLE IF NOT EXISTS batch_changes_site_credentials (
    id BIGSERIAL PRIMARY KEY,
    external_service_type text NOT NULL,
    external_service_id text NOT NULL,
    credential text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX batch_changes_site_credentials_unique ON batch_changes_site_credentials(external_service_type, external_service_id);

COMMIT;
