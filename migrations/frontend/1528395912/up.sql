BEGIN;

-- Create the OOB migration according to doc/dev/background-information/oobmigrations.md
INSERT INTO out_of_band_migrations (id, team, component, description, introduced_version_major, introduced_version_minor, non_destructive)
VALUES (
    12,                                             -- This must be consistent across all Sourcegraph instances
    'apidocs',                                      -- Team owning migration
    'codeintel-db.lsif_data_documentation_search',  -- Component being migrated
    'Index API docs for search',                    -- Description
    3,                                              -- The next minor release (major version)
    32,                                             -- The next minor release (minor version)
    true                                            -- Can be read with previous version without down migration
)
ON CONFLICT DO NOTHING;

COMMIT;
