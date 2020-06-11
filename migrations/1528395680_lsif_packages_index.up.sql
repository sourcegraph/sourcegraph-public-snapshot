BEGIN;

DROP INDEX lsif_packages_package_unique;
CREATE INDEX lsif_packages_scheme_name_version ON lsif_packages (scheme, name, version);

COMMIT;
