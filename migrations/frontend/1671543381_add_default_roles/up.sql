-- read only roles that come with every sourcegraph instance
INSERT INTO 
    roles 
VALUES 
    (1, 'DEFAULT', '2023-01-04 16:29:41.195966+00', NULL, TRUE),
    (2, 'SITE_ADMINISTRATOR', '2023-01-04 16:29:41.195966+00', NULL, TRUE);

SELECT pg_catalog.setval('roles_id_seq', 3, true);
