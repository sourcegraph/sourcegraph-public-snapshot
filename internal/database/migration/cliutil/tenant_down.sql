-- tenant id may not be set in session now, so default to null
ALTER TABLE {{.Table}}
  ALTER COLUMN tenant_id SET DEFAULT NULL,
  ALTER COLUMN tenant_id DROP NOT NULL;

-- Remove RLS
ALTER TABLE {{.Table}} DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS {{.Table}}_isolation_policy ON {{.Table}};
ALTER TABLE {{.Table}} NO FORCE ROW LEVEL SECURITY;
