DELETE FROM external_service_repos WHERE user_id IS NOT NULL OR org_id IS NOT NULL;
DELETE FROM external_services WHERE namespace_user_id IS NOT NULL OR namespace_org_id IS NOT NULL;
