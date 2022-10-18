-- We are changing the schema and we don't expect any rows yet so we can just drop
-- the old table
DROP TABLE IF EXISTS webhooks;

CREATE TABLE webhooks (
      id SERIAL PRIMARY KEY,
      rand_id uuid DEFAULT gen_random_uuid() NOT NULL,
      code_host_kind text NOT NULL,
      code_host_urn text NOT NULL,
      secret text,
      created_at timestamp with time zone DEFAULT now() NOT NULL,
      updated_at timestamp with time zone DEFAULT now() NOT NULL,
      encryption_key_id text
);

COMMENT ON TABLE webhooks IS 'Webhooks registered in Sourcegraph instance.';

COMMENT ON COLUMN webhooks.code_host_kind IS 'Kind of an external service for which webhooks are registered.';

COMMENT ON COLUMN webhooks.code_host_urn IS 'URN of a code host. This column maps to external_service_id column of repo table.';

COMMENT ON COLUMN webhooks.secret IS 'Secret used to decrypt webhook payload (if supported by the code host).';

COMMENT ON COLUMN webhooks.rand_id IS 'rand_id will be the user facing ID';
