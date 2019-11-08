BEGIN;
SET CONSTRAINTS ALL DEFERRED;

WITH batch AS (
  SELECT * FROM ROWS FROM (
  json_to_recordset('[
        {
            "id": 0,
            "repo_id": 1,
            "created_at": "2019-11-11T10:50:36.091729Z",
            "updated_at": "2019-11-11T10:50:36.091729Z",
            "metadata": {},
            "campaign_ids": {},
            "external_id": "foobar-0",
            "external_service_type": "github",
            "campaign_job_id": null,
            "error": null
        },
        {
            "id": 0,
            "repo_id": 2,
            "created_at": "2019-11-11T10:50:36.091729Z",
            "updated_at": "2019-11-11T10:50:36.091729Z",
            "metadata": {},
            "campaign_ids": {},
            "external_id": "foobar-1",
            "external_service_type": "github",
            "campaign_job_id": 2,
            "error": null
        },
        {
            "id": 0,
            "repo_id": 3,
            "created_at": "2019-11-11T10:50:36.091729Z",
            "updated_at": "2019-11-11T10:50:36.091729Z",
            "metadata": {},
            "campaign_ids": {},
            "external_id": null,
            "external_service_type": "github",
            "campaign_job_id": 1,
            "error": null
        }
    ]')
  AS (
      id                    bigint,
      repo_id               integer,
      created_at            timestamptz,
      updated_at            timestamptz,
      metadata              jsonb,
      campaign_ids          jsonb,
      external_id           text,
      external_service_type text,
          campaign_job_id       int,
          error                 text
    )
  )
  WITH ORDINALITY
)
,
-- source: pkg/a8n/store.go:CreateChangesets
changed AS (
  INSERT INTO changesets (
    repo_id,
    created_at,
    updated_at,
    metadata,
    campaign_ids,
    external_id,
    external_service_type,
        campaign_job_id,
        error
  )
  SELECT
    repo_id,
    created_at,
    updated_at,
    metadata,
    campaign_ids,
    external_id,
    external_service_type,
        campaign_job_id,
        error
  FROM batch
  ON CONFLICT (repo_id, external_id, campaign_job_id) WHERE external_id IS NOT NULL AND campaign_job_id IS NOT NULL 
  DO NOTHING
  RETURNING changesets.*
)

SELECT
  COALESCE(changed.id, existing.id) AS id,
  COALESCE(changed.repo_id, existing.repo_id) AS repo_id,
  COALESCE(changed.created_at, existing.created_at) AS created_at,
  COALESCE(changed.updated_at, existing.updated_at) AS updated_at,
  COALESCE(changed.metadata, existing.metadata) AS metadata,
  COALESCE(changed.campaign_ids, existing.campaign_ids) AS campaign_ids,
  COALESCE(changed.external_id, existing.external_id) AS external_id,
  COALESCE(changed.external_service_type, existing.external_service_type) AS external_service_type,
  COALESCE(changed.campaign_job_id, existing.campaign_job_id) AS campaign_job_id,
  COALESCE(changed.error, existing.error) AS error
FROM changed
RIGHT JOIN batch ON batch.repo_id = changed.repo_id
AND (batch.external_id = changed.external_id
OR batch.campaign_job_id = changed.campaign_job_id)
LEFT JOIN changesets existing ON existing.repo_id = batch.repo_id
AND (existing.external_id = batch.external_id
OR existing.campaign_job_id = batch.campaign_job_id)
ORDER BY batch.ordinality;

COMMIT;
