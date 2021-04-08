BEGIN;

-- Add new nullable columns
ALTER TABLE lsif_data_documents ADD COLUMN ranges bytea;
ALTER TABLE lsif_data_documents ADD COLUMN hovers bytea;
ALTER TABLE lsif_data_documents ADD COLUMN monikers bytea;
ALTER TABLE lsif_data_documents ADD COLUMN packages bytea;
ALTER TABLE lsif_data_documents ADD COLUMN diagnostics bytea;

-- Add new comments
COMMENT ON COLUMN lsif_data_documents.ranges IS 'A gob-encoded payload conforming to the [Ranges](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L14:2) field of the DocumentDatatype.';
COMMENT ON COLUMN lsif_data_documents.hovers IS 'A gob-encoded payload conforming to the [HoversResults](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L15:2) field of the DocumentDatatype.';
COMMENT ON COLUMN lsif_data_documents.monikers IS 'A gob-encoded payload conforming to the [Monikers](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L16:2) field of the DocumentDatatype.';
COMMENT ON COLUMN lsif_data_documents.packages IS 'A gob-encoded payload conforming to the [PackageInformation](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L17:2) field of the DocumentDatatype.';
COMMENT ON COLUMN lsif_data_documents.diagnostics IS 'A gob-encoded payload conforming to the [Diagnostics](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L18:2) field of the DocumentDatatype.';

-- Update outdated comments
COMMENT ON COLUMN lsif_data_documents.data IS 'A gob-encoded payload conforming to the [DocumentData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L13:6) type. This field is being migrated across ranges, hovers, monikers, packages, and diagnostics columns and will be removed in a future release of Sourcegraph.';
COMMENT ON COLUMN lsif_data_result_chunks.data IS 'A gob-encoded payload conforming to the [ResultChunkData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L76:6) type.';
COMMENT ON COLUMN lsif_data_definitions.data IS 'A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.';
COMMENT ON COLUMN lsif_data_references.data IS 'A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.';

COMMIT;
