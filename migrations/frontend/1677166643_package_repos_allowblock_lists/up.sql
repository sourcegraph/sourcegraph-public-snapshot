CREATE TABLE IF NOT EXISTS package_repo_filters (
    id SERIAL PRIMARY KEY NOT NULL,
    behaviour TEXT NOT NULL,
    external_service_kind TEXT NOT NULL,
    matcher JSONB NOT NULL,
    -- because creating types is unnecessarily awkward with idempotency
    CONSTRAINT kind_is_package_repo_kind CHECK (
        external_service_kind = ANY('{"JVMPACKAGES","NPMPACKAGES","GOPACKAGES","PYTHONPACKAGES","RUSTPACKAGES","RUBYPACKAGES"}'),
    )
    CONSTRAINT behaviour_is_allow_or_block CHECK (
        behaviour = ANY('{"BLOCK","ALLOW"}'),
    )
);

CREATE INDEX IF NOT EXISTS package_repo_allowblock_list_extsvc_kind
ON package_repo_allowblock_list USING btree (external_service_kind);
