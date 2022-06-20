CREATE INDEX IF NOT EXISTS
    changesets_bitbucket_cloud_metadata_source_commit_idx
ON
    changesets ((metadata->'source'->'commit'->>'hash'))
