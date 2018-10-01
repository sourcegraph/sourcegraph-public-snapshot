
ALTER TABLE phabricator_repos DROP COLUMN url;

ALTER TABLE phabricator_repos
	ADD UNIQUE (callsign);

COMMIT;