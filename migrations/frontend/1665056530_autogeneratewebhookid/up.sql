ALTER TABLE IF EXISTS webhooks ALTER COLUMN id SET DEFAULT gen_random_uuid();

ALTER TABLE IF EXISTS webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);
