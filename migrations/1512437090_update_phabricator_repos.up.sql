
ALTER TABLE phabricator_repos ADD COLUMN url TEXT NOT NULL DEFAULT '';

ALTER TABLE phabricator_repos
	DROP CONSTRAINT phabricator_repos_callsign_key;

COMMIT;