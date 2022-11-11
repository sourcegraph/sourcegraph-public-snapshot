-- Undo the changes made in the up migration
ALTER TABLE gitserver_repos
ADD repo_status text;
