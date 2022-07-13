UPDATE lsif_dependency_repos
SET name = replace(regexp_replace(name, '^maven/', ''), '/', ':')
WHERE scheme = 'semanticdb'
AND name LIKE 'maven/%';

DELETE FROM lsif_dependency_repos
WHERE scheme = 'semanticdb'
AND (name LIKE '%:%:%' OR name LIKE 'jdk:%' OR name LIKE 'jdk/%');
