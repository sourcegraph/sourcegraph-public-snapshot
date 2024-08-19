-- unwritten columns become default tenant_id
UPDATE {{.Table}} SET tenant_id = 1 WHERE tenant_id IS NULL;

-- future inserts will require the session tenant_id
ALTER TABLE {{.Table}}
  ALTER COLUMN tenant_id SET DEFAULT current_setting('app.current_tenant')::integer,
  ALTER COLUMN tenant_id SET NOT NULL;

-- you can only read an entry if your session tenant_id matches
ALTER TABLE {{.Table}} ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS {{.Table}}_isolation_policy ON {{.Table}};
CREATE POLICY {{.Table}}_isolation_policy ON {{.Table}} USING (tenant_id = current_setting('app.current_tenant')::integer);
ALTER TABLE {{.Table}} FORCE ROW LEVEL SECURITY;
