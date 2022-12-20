-- read only roles that come with every sourcegraph instance
INSERT INTO roles VALUES (1, 'DEFAULT', NOW(), NULL, TRUE);
INSERT INTO roles VALUES (2, 'SITE_ADMINISTRATOR', NOW(), NULL, TRUE);
