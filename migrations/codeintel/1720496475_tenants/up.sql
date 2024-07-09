-- todo: no tenants table in the codeintel DB so need to replicate here..

CREATE TABLE IF NOT EXISTS tenants (
    id SERIAL PRIMARY KEY,
    name text NOT NULL,
    slug text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

INSERT INTO tenants (id, name, slug, created_at, updated_at) VALUES (1, 'default', 'default', NOW(), NOW()) ON CONFLICT (id) DO NOTHING;
SET app.current_tenant = 1;

-- Temporary function to deduplicate the above queries for each table:
CREATE OR REPLACE FUNCTION migrate_table(table_name text)
RETURNS void AS $$
BEGIN
    -- todo: non nullable column?
    EXECUTE format('
        ALTER TABLE %I ADD COLUMN IF NOT EXISTS tenant_id integer DEFAULT current_setting(''app.current_tenant'')::integer REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE;
        -- TODO REMOVE
        UPDATE %I SET tenant_id = 1;
        ALTER TABLE %I ALTER COLUMN tenant_id SET NOT NULL;
        ALTER TABLE %I ENABLE ROW LEVEL SECURITY;
        DROP POLICY IF EXISTS %I_isolation_policy ON %I;
        CREATE POLICY %I_isolation_policy ON %I USING (tenant_id = current_setting(''app.current_tenant'')::integer);
        ALTER TABLE %I FORCE ROW LEVEL SECURITY;',
        table_name, table_name, table_name, table_name, table_name, table_name, table_name, table_name, table_name
    );
END;
$$ LANGUAGE plpgsql;

-- todo, those might all need to run in separate transactions, as pending triggers can fail here.
SELECT migrate_table('rockskip_ancestry');
SELECT migrate_table('rockskip_repos');
SELECT migrate_table('rockskip_symbols');

DROP FUNCTION migrate_table(text);
