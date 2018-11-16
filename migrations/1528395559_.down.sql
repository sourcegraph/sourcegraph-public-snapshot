ALTER TABLE registry_extension_releases ALTER COLUMN manifest TYPE text USING manifest#>>'{}';
