BEGIN;

ALTER TABLE external_services
ADD COLUMN IF NOT EXISTS
    cloud_global BOOLEAN DEFAULT false;

-- Only only service per kind can have cloud_global set
CREATE UNIQUE INDEX IF NOT EXISTS kind_cloud_global ON external_services (kind, cloud_global) WHERE cloud_global = true;

COMMIT;
