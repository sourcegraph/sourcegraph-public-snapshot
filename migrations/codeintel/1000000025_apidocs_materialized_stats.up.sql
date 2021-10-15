BEGIN;

---------------------------------------------------------------------------
-- Materialized stats views for the lsif_data_documentation_pages table. --
---------------------------------------------------------------------------
CREATE MATERIALIZED VIEW lsif_data_apidocs_num_pages
AS SELECT count(*) FROM lsif_data_documentation_pages
WITH DATA;

CREATE MATERIALIZED VIEW lsif_data_apidocs_num_dumps
AS SELECT count(DISTINCT dump_id) FROM lsif_data_documentation_pages
WITH DATA;

-- Materialized view for reporting progress of our OOB migration, which is expensive
-- to calculate even once the migration has succeeded.
CREATE MATERIALIZED VIEW lsif_data_documentation_pages_oob_migrated
AS SELECT
    CASE c2.count WHEN 0 THEN 1 ELSE cast(c1.count as float) / cast(c2.count as float) END AS percent
    FROM
        (SELECT count(DISTINCT dump_id) FROM lsif_data_documentation_pages WHERE search_indexed='true') c1,
        (SELECT count(DISTINCT dump_id) FROM lsif_data_documentation_pages) c2
WITH DATA;

CREATE OR REPLACE FUNCTION refresh_lsif_data_documentation_pages()
RETURNS TRIGGER LANGUAGE plpgsql
AS $$
BEGIN
REFRESH MATERIALIZED VIEW CONCURRENTLY lsif_data_apidocs_num_pages;
REFRESH MATERIALIZED VIEW CONCURRENTLY lsif_data_apidocs_num_dumps;
REFRESH MATERIALIZED VIEW CONCURRENTLY lsif_data_documentation_pages_oob_migrated;
RETURN NULL;
END $$;

CREATE TRIGGER refresh_lsif_data_documentation_pages
AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
ON lsif_data_documentation_pages
FOR EACH STATEMENT
EXECUTE PROCEDURE refresh_lsif_data_documentation_pages();

--------------------------------------------------------------------------
-- Materialized stats views for the lsif_data_docs_search_public table. --
--------------------------------------------------------------------------
CREATE MATERIALIZED VIEW lsif_data_apidocs_num_search_results_public
AS SELECT count(*) FROM lsif_data_docs_search_public
WITH DATA;

CREATE OR REPLACE FUNCTION refresh_lsif_data_docs_search_public()
RETURNS TRIGGER LANGUAGE plpgsql
AS $$
BEGIN
REFRESH MATERIALIZED VIEW CONCURRENTLY lsif_data_apidocs_num_search_results_public;
RETURN NULL;
END $$;

CREATE TRIGGER refresh_lsif_data_docs_search_public
AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
ON lsif_data_docs_search_public
FOR EACH STATEMENT
EXECUTE PROCEDURE refresh_lsif_data_docs_search_public();

---------------------------------------------------------------------------
-- Materialized stats views for the lsif_data_docs_search_private table. --
---------------------------------------------------------------------------
CREATE MATERIALIZED VIEW lsif_data_apidocs_num_search_results_private
AS SELECT count(*) FROM lsif_data_docs_search_private
WITH DATA;

CREATE OR REPLACE FUNCTION refresh_lsif_data_docs_search_private()
RETURNS TRIGGER LANGUAGE plpgsql
AS $$
BEGIN
REFRESH MATERIALIZED VIEW CONCURRENTLY lsif_data_apidocs_num_search_results_private;
RETURN NULL;
END $$;

CREATE TRIGGER refresh_lsif_data_docs_search_private
AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
ON lsif_data_docs_search_private
FOR EACH STATEMENT
EXECUTE PROCEDURE refresh_lsif_data_docs_search_private();

COMMIT;
