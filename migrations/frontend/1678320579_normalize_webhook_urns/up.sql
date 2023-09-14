UPDATE webhooks
-- Add a trailing slash to URNs that don't have one. This mimics the effect of the
-- extsvc.NormalizeBaseURL method in Go
SET code_host_urn = CONCAT(code_host_urn, '/')
WHERE code_host_urn NOT LIKE '%/';
