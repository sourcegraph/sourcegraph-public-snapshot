CREATE TABLE IF NOT EXISTS webhooks(
    id UUID NOT NULL,
    code_host_kind TEXT NOT NULL,
    code_host_urn TEXT NOT NULL,
    secret TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    encryption_key_id TEXT NOT NULL DEFAULT ''
);

COMMENT ON TABLE webhooks IS 'Webhooks registered in Sourcegraph instance.';
COMMENT ON COLUMN webhooks.code_host_kind IS 'Kind of an external service for which webhooks are registered.';
COMMENT ON COLUMN webhooks.code_host_urn IS 'URN of a code host. This column maps to external_service_id column of repo table.';
COMMENT ON COLUMN webhooks.secret IS 'Secret used to decrypt webhook payload (if supported by the code host).';
