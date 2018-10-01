ALTER TABLE repo
  ADD COLUMN vcs text,
  ADD COLUMN http_clone_url text,
  ADD COLUMN ssh_clone_url text,
  ADD COLUMN homepage_url text,
  ADD COLUMN default_branch text,
  ADD COLUMN deprecated boolean,
  ADD COLUMN mirror boolean,
  ADD COLUMN vcs_synced_at timestamp with time zone;
