UPDATE users SET external_provider='auth0' WHERE external_id LIKE 'auth0|%';
UPDATE users SET external_provider=null WHERE external_provider IN ('', 'native');
UPDATE users SET external_id=null WHERE external_id='' OR external_id ILIKE 'native:%';
ALTER TABLE users ADD CONSTRAINT check_external_id CHECK ((external_provider IS NULL) = (external_id IS NULL));
