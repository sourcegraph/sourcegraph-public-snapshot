ALTER TABLE webhooks
   DROP CONSTRAINT IF EXISTS webhooks_pkey,
   ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id),
   ALTER COLUMN id SET DEFAULT gen_random_uuid();
