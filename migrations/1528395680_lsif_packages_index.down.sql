BEGIN;

DROP INDEX lsif_packages_scheme_name_version;
CREATE UNIQUE INDEX lsif_packages_package_unique ON lsif_packages (scheme, name, version);

COMMIT;
