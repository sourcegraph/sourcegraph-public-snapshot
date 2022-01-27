BEGIN;

--------------------------------------------------------
-- Stats for the lsif_data_documentation_pages table. --
--------------------------------------------------------
CREATE TABLE lsif_data_apidocs_num_pages AS SELECT count(*) FROM lsif_data_documentation_pages;
CREATE TABLE lsif_data_apidocs_num_dumps AS SELECT count(DISTINCT dump_id) FROM lsif_data_documentation_pages;
CREATE TABLE lsif_data_apidocs_num_dumps_indexed AS SELECT count(DISTINCT dump_id) FROM lsif_data_documentation_pages WHERE search_indexed='true';

CREATE OR REPLACE FUNCTION lsif_data_documentation_pages_delete()
RETURNS TRIGGER LANGUAGE plpgsql
AS $$
BEGIN
-- Decrement tally counting tables.
UPDATE lsif_data_apidocs_num_pages SET count = count - (select count(*) from oldtbl);
UPDATE lsif_data_apidocs_num_dumps SET count = count - (select count(DISTINCT dump_id) from oldtbl);
UPDATE lsif_data_apidocs_num_dumps_indexed SET count = count - (select count(DISTINCT dump_id) from oldtbl WHERE search_indexed='true');
RETURN NULL;
END $$;

CREATE TRIGGER lsif_data_documentation_pages_delete
AFTER DELETE ON lsif_data_documentation_pages
REFERENCING OLD TABLE AS oldtbl
FOR EACH STATEMENT EXECUTE PROCEDURE lsif_data_documentation_pages_delete();

CREATE OR REPLACE FUNCTION lsif_data_documentation_pages_insert()
RETURNS TRIGGER LANGUAGE plpgsql
AS $$
BEGIN
-- Increment tally counting tables.
UPDATE lsif_data_apidocs_num_pages SET count = count + (select count(*) from newtbl);
UPDATE lsif_data_apidocs_num_dumps SET count = count + (select count(DISTINCT dump_id) from newtbl);
UPDATE lsif_data_apidocs_num_dumps_indexed SET count = count + (select count(DISTINCT dump_id) from newtbl WHERE search_indexed='true');
RETURN NULL;
END $$;

CREATE TRIGGER lsif_data_documentation_pages_insert
AFTER INSERT ON lsif_data_documentation_pages
REFERENCING NEW TABLE AS newtbl
FOR EACH STATEMENT EXECUTE PROCEDURE lsif_data_documentation_pages_insert();

CREATE OR REPLACE FUNCTION lsif_data_documentation_pages_update()
RETURNS TRIGGER LANGUAGE plpgsql
AS $$
BEGIN
WITH
    beforeIndexed AS (SELECT count(DISTINCT dump_id) FROM oldtbl WHERE search_indexed='true'),
    afterIndexed AS (SELECT count(DISTINCT dump_id) FROM newtbl WHERE search_indexed='true')
UPDATE lsif_data_apidocs_num_dumps_indexed SET count=count + ((select * from afterIndexed) - (select * from beforeIndexed));
RETURN NULL;
END $$;

CREATE TRIGGER lsif_data_documentation_pages_update
AFTER UPDATE ON lsif_data_documentation_pages
REFERENCING OLD TABLE AS oldtbl NEW TABLE AS newtbl
FOR EACH STATEMENT EXECUTE PROCEDURE lsif_data_documentation_pages_update();

-------------------------------------------------------
-- Stats for the lsif_data_docs_search_public table. --
-------------------------------------------------------
CREATE TABLE lsif_data_apidocs_num_search_results_public AS SELECT count(*) FROM lsif_data_docs_search_public;

CREATE OR REPLACE FUNCTION lsif_data_docs_search_public_delete()
RETURNS TRIGGER LANGUAGE plpgsql
AS $$
BEGIN
-- Decrement tally counting tables.
UPDATE lsif_data_apidocs_num_search_results_public SET count = count - (select count(*) from oldtbl);
RETURN NULL;
END $$;

CREATE TRIGGER lsif_data_docs_search_public_delete
AFTER DELETE
ON lsif_data_docs_search_public
REFERENCING OLD TABLE AS oldtbl
FOR EACH STATEMENT EXECUTE PROCEDURE lsif_data_docs_search_public_delete();

CREATE OR REPLACE FUNCTION lsif_data_docs_search_public_insert()
RETURNS TRIGGER LANGUAGE plpgsql
AS $$
BEGIN
-- Increment tally counting tables.
UPDATE lsif_data_apidocs_num_search_results_public SET count = count + (select count(*) from newtbl);
RETURN NULL;
END $$;

CREATE TRIGGER lsif_data_docs_search_public_insert
AFTER INSERT
ON lsif_data_docs_search_public
REFERENCING NEW TABLE AS newtbl
FOR EACH STATEMENT EXECUTE PROCEDURE lsif_data_docs_search_public_insert();

-------------------------------------------------------
-- Stats for the lsif_data_docs_search_private table. --
-------------------------------------------------------
CREATE TABLE lsif_data_apidocs_num_search_results_private AS SELECT count(*) FROM lsif_data_docs_search_private;

CREATE OR REPLACE FUNCTION lsif_data_docs_search_private_delete()
RETURNS TRIGGER LANGUAGE plpgsql
AS $$
BEGIN
-- Decrement tally counting tables.
UPDATE lsif_data_apidocs_num_search_results_private SET count = count - (select count(*) from oldtbl);
RETURN NULL;
END $$;

CREATE TRIGGER lsif_data_docs_search_private_delete
AFTER DELETE
ON lsif_data_docs_search_private
REFERENCING OLD TABLE AS oldtbl
FOR EACH STATEMENT EXECUTE PROCEDURE lsif_data_docs_search_private_delete();

CREATE OR REPLACE FUNCTION lsif_data_docs_search_private_insert()
RETURNS TRIGGER LANGUAGE plpgsql
AS $$
BEGIN
-- Increment tally counting tables.
UPDATE lsif_data_apidocs_num_search_results_private SET count = count + (select count(*) from newtbl);
RETURN NULL;
END $$;

CREATE TRIGGER lsif_data_docs_search_private_insert
AFTER INSERT
ON lsif_data_docs_search_private
REFERENCING NEW TABLE AS newtbl
FOR EACH STATEMENT EXECUTE PROCEDURE lsif_data_docs_search_private_insert();

COMMIT;
