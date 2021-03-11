BEGIN;

COMMENT ON TABLE lsif_uploads IS 'Stores metadata about an LSIF index uploaded by a user.';
COMMENT ON COLUMN lsif_uploads.id IS 'Used as a logical foreign key with the (disjoint) codeintel database.';
COMMENT ON COLUMN lsif_uploads.commit IS 'A 40-char revhash. Note that this commit may not be resolvable in the future.';
COMMENT ON COLUMN lsif_uploads.root IS 'The path for which the index can resolve code intelligence relative to the repository root.';
COMMENT ON COLUMN lsif_uploads.indexer IS 'The name of the indexer that produced the index file. If not supplied by the user it will be pulled from the index metadata.';
COMMENT ON COLUMN lsif_uploads.num_parts IS 'The number of parts src-cli split the upload file into.';
COMMENT ON COLUMN lsif_uploads.uploaded_parts IS 'The index of parts that have been successfully uploaded.';
COMMENT ON COLUMN lsif_uploads.upload_size IS 'The size of the index file (in bytes).';
-- Omitting repository_id
-- Commenting on these in another PR: state, uploaded_at, started_at, finished_at, process_after, num_resets, num_failures, failure_message

COMMENT ON TABLE lsif_packages IS 'Associates an upload with the set of packages they provide within a given packages management scheme.';
COMMENT ON COLUMN lsif_packages.scheme IS 'The (export) moniker scheme.';
COMMENT ON COLUMN lsif_packages.name IS 'The package name.';
COMMENT ON COLUMN lsif_packages.version IS 'The package version.';
COMMENT ON COLUMN lsif_packages.dump_id IS 'The identifier of the upload that provides the package.';
-- Omitting id

COMMENT ON TABLE lsif_references IS 'Associates an upload with the set of packages they require within a given packages management scheme.';
COMMENT ON COLUMN lsif_references.scheme IS 'The (import) moniker scheme.';
COMMENT ON COLUMN lsif_references.name IS 'The package name.';
COMMENT ON COLUMN lsif_references.version IS 'The package version.';
COMMENT ON COLUMN lsif_references.filter IS 'A [bloom filter](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/enterprise/internal/codeintel/bloomfilter/bloom_filter.go#L27:6) encoded as gzipped JSON. This bloom filter stores the set of identifiers imported from the package.';
COMMENT ON COLUMN lsif_references.dump_id IS 'The identifier of the upload that references the package.';
-- Omitting id

COMMENT ON TABLE lsif_nearest_uploads IS 'Associates commits with the complete set of uploads visible from that commit. Every commit with upload data is present in this table.';
COMMENT ON COLUMN lsif_nearest_uploads.commit_bytea IS 'A 40-char revhash. Note that this commit may not be resolvable in the future.';
COMMENT ON COLUMN lsif_nearest_uploads.uploads IS 'Encodes an {upload_id => distance} map that includes an entry for every upload visible from the commit. There is always at least one entry with a distance of zero.';
-- Omitting repository_id

COMMENT ON TABLE lsif_nearest_uploads_links IS 'Associates commits with the closest ancestor commit with usable upload data. Together, this table and lsif_nearest_uploads cover all commits with resolvable code intelligence.';
COMMENT ON COLUMN lsif_nearest_uploads_links.commit_bytea IS 'A 40-char revhash. Note that this commit may not be resolvable in the future.';
COMMENT ON COLUMN lsif_nearest_uploads_links.ancestor_commit_bytea IS 'The 40-char revhash of the ancestor. Note that this commit may not be resolvable in the future.';
COMMENT ON COLUMN lsif_nearest_uploads_links.distance IS 'The distance bewteen the commits. Parent = 1, Grandparent = 2, etc.';
-- Omitting repository_id

COMMENT ON TABLE lsif_uploads_visible_at_tip IS 'Associates a repository with the set of LSIF upload identifiers that can serve intelligence for the tip of the default branch.';
COMMENT ON COLUMN lsif_uploads_visible_at_tip.upload_id IS 'The identifier of an upload visible at the tip of the default branch.';
-- Omitting repository_id

COMMENT ON TABLE lsif_dirty_repositories IS 'Stores whether or not the nearest upload data for a repository is out of date (when update_token > dirty_token).';
COMMENT ON COLUMN lsif_dirty_repositories.update_token IS 'This value is incremented on each request to update the commit graph for the repository.';
COMMENT ON COLUMN lsif_dirty_repositories.dirty_token IS 'Set to the value of update_token visible to the transaction that updates the commit graph. Updates of dirty_token during this time will cause a second update.';
-- Omitting repository_id

COMMENT ON TABLE lsif_indexes IS 'Stores metadata about a code intel index job.';
COMMENT ON COLUMN lsif_indexes.commit IS 'A 40-char revhash. Note that this commit may not be resolvable in the future.';
COMMENT ON COLUMN lsif_indexes.docker_steps IS 'An array of pre-index [steps](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/enterprise/internal/codeintel/stores/dbstore/docker_step.go#L9:6) to run.';
COMMENT ON COLUMN lsif_indexes.indexer IS 'The docker image used to run the index command (e.g. sourcegraph/lsif-go).';
COMMENT ON COLUMN lsif_indexes.root IS 'The working directory of the indexer image relative to the repository root.';
COMMENT ON COLUMN lsif_indexes.local_steps IS 'A list of commands to run inside the indexer image prior to running the indexer command.';
COMMENT ON COLUMN lsif_indexes.indexer_args IS 'The command run inside the indexer image to produce the index file (e.g. [''lsif-node'', ''-p'', ''.''])';
COMMENT ON COLUMN lsif_indexes.outfile IS 'The path to the index file produced by the index command relative to the working directory.';

COMMENT ON COLUMN lsif_indexes.execution_logs IS 'An array of [log entries](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/internal/workerutil/store.go#L48:6) (encoded as JSON) from the most recent execution.';
COMMENT ON COLUMN lsif_indexes.log_contents IS '**Column deprecated in favor of execution_logs.**';
-- Omitting id, repository_id
-- Commenting on these in another PR: state, queued_at, started_at, finished_at, process_after, num_resets, num_failures, failure_message

COMMENT ON TABLE lsif_indexable_repositories IS 'Stores the number of code intel events for repositories. Used for auto-index scheduling heursitics Sourcegraph Cloud.';
COMMENT ON COLUMN lsif_indexable_repositories.search_count IS 'The number of search-based code intel events for the repository in the past week.';
COMMENT ON COLUMN lsif_indexable_repositories.precise_count IS 'The number of precise code intel events for the repository in the past week.';
COMMENT ON COLUMN lsif_indexable_repositories.last_index_enqueued_at IS 'The last time an index for the repository was enqueued (for basic rate limiting).';
COMMENT ON COLUMN lsif_indexable_repositories.last_updated_at IS 'The last time the event counts were updated for this repository.';
COMMENT ON COLUMN lsif_indexable_repositories.enabled IS '**Column unused.**';
-- Omitting id, repository_id

COMMENT ON TABLE lsif_index_configuration IS 'Stores the configuration used for code intel index jobs for a repository.';
COMMENT ON COLUMN lsif_index_configuration.data IS 'The raw user-supplied [configuration](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.23/-/blob/enterprise/internal/codeintel/autoindex/config/types.go#L3:6) (encoded in JSONC).';
-- Omitting id, repository_id

COMMIT;
