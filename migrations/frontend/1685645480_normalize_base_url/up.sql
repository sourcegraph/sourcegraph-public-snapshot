UPDATE github_apps
-- Add a trailing slash to URNs that don't have one. This mimics the effect of the
-- extsvc.NormalizeBaseURL method in Go
SET base_url = CONCAT(base_url, '/')
WHERE base_url NOT LIKE '%/';
