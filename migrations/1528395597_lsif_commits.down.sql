-- The data here has been removed for new customers before the release of 3.9.
-- We originally tried to put LSIF data in a separate database, which did not
-- work as well as we'd hoped before it got into master. This migration file
-- is staying here so that it does not disturb developers and CI/CDd instances
-- of Sourcegraph.

-- The effective behavior of this migration has moved to the following pair of
-- migration files:
--   * 1528395599_create_lsif_tables.up.sql.
--   * 1528395599_create_lsif_tables.down.sql.

SELECT 1;
