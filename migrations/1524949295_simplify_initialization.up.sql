-- If any sites were in a state where users were created (e.g., from SSO) but the site was not
-- initialized, there is no way for site admins to recover without resorting to executing SQL
-- queries manually.
--
-- This statement extricates those sites from that bad state. It is safe because initialized=true
-- does not expose any additional data and it makes user account creation strictly stricter (in that
-- an account is not made site admin if there are no other accounts).
UPDATE site_config SET initialized=true WHERE EXISTS(SELECT * FROM users);
