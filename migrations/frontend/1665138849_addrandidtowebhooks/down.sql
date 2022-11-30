-- We can just drop the new table and create the old one again

DROP TABLE IF EXISTS webhooks;

CREATE TABLE webhooks (
      id uuid DEFAULT gen_random_uuid() NOT NULL,
      code_host_kind text NOT NULL,
      code_host_urn text NOT NULL,
      secret text,
      created_at timestamp with time zone DEFAULT now() NOT NULL,
      updated_at timestamp with time zone DEFAULT now() NOT NULL,
      encryption_key_id text DEFAULT ''::text NOT NULL
);

COMMENT ON TABLE webhooks IS 'Webhooks registered in Sourcegraph instance.';

COMMENT ON COLUMN webhooks.code_host_kind IS 'Kind of an external service for which webhooks are registered.';

COMMENT ON COLUMN webhooks.code_host_urn IS 'URN of a code host. This column maps to external_service_id column of repo table.';

COMMENT ON COLUMN webhooks.secret IS 'Secret used to decrypt webhook payload (if supported by the code host).';

ALTER TABLE webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);
