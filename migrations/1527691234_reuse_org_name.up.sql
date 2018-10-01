-- Soft-deleted orgs (deleted_at IS NOT NULL) should not be accounted for in unique indexes.

ALTER TABLE orgs DROP CONSTRAINT org_name_unique;
CREATE UNIQUE INDEX orgs_name ON orgs(name) WHERE deleted_at IS NULL;
