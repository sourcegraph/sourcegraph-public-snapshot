DROP TRIGGER IF EXISTS lsif_data_docs_search_private_delete ON lsif_data_docs_search_private;

DROP TRIGGER IF EXISTS lsif_data_docs_search_private_insert ON lsif_data_docs_search_private;

DROP TRIGGER IF EXISTS lsif_data_docs_search_public_delete ON lsif_data_docs_search_public;

DROP TRIGGER IF EXISTS lsif_data_docs_search_public_insert ON lsif_data_docs_search_public;

DROP TRIGGER IF EXISTS lsif_data_documentation_pages_delete ON lsif_data_documentation_pages;

DROP TRIGGER IF EXISTS lsif_data_documentation_pages_insert ON lsif_data_documentation_pages;

DROP TRIGGER IF EXISTS lsif_data_documentation_pages_update ON lsif_data_documentation_pages;

DROP FUNCTION IF EXISTS lsif_data_docs_search_private_delete();

DROP FUNCTION IF EXISTS lsif_data_docs_search_private_insert();

DROP FUNCTION IF EXISTS lsif_data_docs_search_public_delete();

DROP FUNCTION IF EXISTS lsif_data_docs_search_public_insert();

DROP FUNCTION IF EXISTS lsif_data_documentation_pages_delete();

DROP FUNCTION IF EXISTS lsif_data_documentation_pages_insert();

DROP FUNCTION IF EXISTS lsif_data_documentation_pages_update();

DROP TABLE IF EXISTS lsif_data_apidocs_num_dumps CASCADE;

DROP TABLE IF EXISTS lsif_data_apidocs_num_dumps_indexed CASCADE;

DROP TABLE IF EXISTS lsif_data_apidocs_num_pages CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_apidocs_num_pages_id_seq;

DROP TABLE IF EXISTS lsif_data_apidocs_num_search_results_private CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_apidocs_num_search_results_private_id_seq;

DROP TABLE IF EXISTS lsif_data_apidocs_num_search_results_public CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_apidocs_num_search_results_public_id_seq;

DROP TABLE IF EXISTS lsif_data_docs_search_current_private CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_docs_search_current_private_id_seq;

DROP TABLE IF EXISTS lsif_data_docs_search_current_public CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_docs_search_current_public_id_seq;

DROP TABLE IF EXISTS lsif_data_docs_search_lang_names_private CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_docs_search_lang_names_private_id_seq;

DROP TABLE IF EXISTS lsif_data_docs_search_lang_names_public CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_docs_search_lang_names_public_id_seq;

DROP TABLE IF EXISTS lsif_data_docs_search_private CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_docs_search_private_id_seq;

DROP TABLE IF EXISTS lsif_data_docs_search_public CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_docs_search_public_id_seq;

DROP TABLE IF EXISTS lsif_data_docs_search_repo_names_private CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_docs_search_repo_names_private_id_seq;

DROP TABLE IF EXISTS lsif_data_docs_search_repo_names_public CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_docs_search_repo_names_public_id_seq;

DROP TABLE IF EXISTS lsif_data_docs_search_tags_private CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_docs_search_tags_private_id_seq;

DROP TABLE IF EXISTS lsif_data_docs_search_tags_public CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_docs_search_tags_public_id_seq;

DROP TABLE IF EXISTS lsif_data_documentation_mappings CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_documentation_mappings_id_seq;

DROP TABLE IF EXISTS lsif_data_documentation_pages CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_documentation_pages_id_seq;

DROP TABLE IF EXISTS lsif_data_documentation_path_info CASCADE;
DROP SEQUENCE IF EXISTS lsif_data_documentation_path_info_id_seq;
