INSERT INTO user_tags (user_id, name) (
	SELECT DISTINCT ON (users.id) users.id, 'editor-beta' FROM org_members INNER JOIN users ON users.auth0_id = org_members.user_id
);

INSERT INTO org_tags (org_id, name) (
	SELECT DISTINCT ON (orgs.id) orgs.id, 'editor-beta' FROM orgs
);
