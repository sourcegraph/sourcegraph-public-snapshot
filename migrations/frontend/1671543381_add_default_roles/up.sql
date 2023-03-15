-- system roles that come with every sourcegraph instance
INSERT INTO
    roles (id, name, created_at, deleted_at, readonly)
VALUES
    (1, 'USER', '2023-01-04 16:29:41.195966+00', NULL, TRUE),
    (2, 'SITE_ADMINISTRATOR', '2023-01-04 16:29:41.195966+00', NULL, TRUE)
ON CONFLICT (id) DO NOTHING;

SELECT pg_catalog.setval('roles_id_seq', 3, true);
