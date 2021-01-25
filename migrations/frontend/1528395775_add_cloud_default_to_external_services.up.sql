BEGIN;

ALTER TABLE external_services
ADD COLUMN IF NOT EXISTS
    cloud_default BOOLEAN DEFAULT false;

-- Only only service per kind can have cloud_default set
CREATE UNIQUE INDEX IF NOT EXISTS kind_cloud_default ON external_services (kind, cloud_default) WHERE cloud_default = true;

COMMIT;
