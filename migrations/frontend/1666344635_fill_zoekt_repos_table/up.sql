INSERT INTO zoekt_repos (repo_id)
SELECT id
FROM repo
LEFT JOIN zoekt_repos zr ON repo.id = zr.repo_id
WHERE zr.repo_id IS NULL
ON CONFLICT (repo_id) DO NOTHING;
