-- shrimple as that
DROP TABLE IF EXISTS package_repo_filters CASCADE;

DROP FUNCTION IF EXISTS glob_to_regex(text);
DROP FUNCTION IF EXISTS is_unversioned_package_allowed(text, text);
DROP FUNCTION IF EXISTS is_versioned_package_allowed(text, text, text);
