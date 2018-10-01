DROP INDEX orgs_name;
ALTER TABLE orgs ADD CONSTRAINT org_name_unique UNIQUE(name);
