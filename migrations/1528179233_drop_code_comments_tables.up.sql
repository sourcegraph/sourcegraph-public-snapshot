-- Drop code comments tables. Use IF EXISTS because these tables are NOT recreated in the
-- downmigration (because the intent is to remove them completely, and they have been unused for a
-- while).
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS threads;
DROP TABLE IF EXISTS org_repos;
DROP TABLE IF EXISTS shared_items;

-- Drop old backup tables. In case admins manually deleted these thinking they were not backups
-- managed by Sourcegraph, use IF EXISTS to prevent failures.
DROP TABLE IF EXISTS comments_bkup_1514545501;
DROP TABLE IF EXISTS shared_items_bkup_1514546912;
DROP TABLE IF EXISTS threads_bkup_1514544774;
