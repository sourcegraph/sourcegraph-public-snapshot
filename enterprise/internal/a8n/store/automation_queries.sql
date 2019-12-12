-- automation_query.sql

-- name: CountChangesets :one
SELECT COUNT(id)
FROM changesets
WHERE campaign_ids ? $1;

-- name: GetChangeset :one
SELECT
  id,
  repo_id,
  created_at,
  updated_at,
  metadata,
  campaign_ids,
  external_id,
  external_service_type
FROM changesets
WHERE id = $1
LIMIT 1;

-- name: ListChangesets :many
SELECT
  id,
  repo_id,
  created_at,
  updated_at,
  metadata,
  campaign_ids,
  external_id,
  external_service_type
FROM changesets
WHERE campaign_ids ? $1
ORDER BY id ASC
LIMIT $2;

-- name: GetChangesetEvent :one
SELECT
    id,
    changeset_id,
    kind,
    key,
    created_at,
    updated_at,
    metadata
FROM changeset_events
WHERE id = $1
LIMIT 1;

-- name: ListChangesetEvents :many
SELECT
    id,
    changeset_id,
    kind,
    key,
    created_at,
    updated_at,
    metadata
FROM changeset_events
WHERE id = $1
ORDER BY id ASC;

-- name: CountChangesetEvents :one
SELECT COUNT(*)
FROM changeset_events
WHERE changeset_id = $1;

-- name: CreateCampaign :one
INSERT INTO campaigns (
  name,
  description,
  author_id,
  namespace_user_id,
  namespace_org_id,
  created_at,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING
  id,
  name,
  description,
  author_id,
  namespace_user_id,
  namespace_org_id,
  created_at,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at;

-- name: UpdateCampaign :one
UPDATE campaigns
SET (
  name,
  description,
  author_id,
  namespace_user_id,
  namespace_org_id,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
) = ($1, $2, $3, $4, $5, $6, $7, $8, $9)
WHERE id = $10
RETURNING
  id,
  name,
  description,
  author_id,
  namespace_user_id,
  namespace_org_id,
  created_at,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at;

-- name: DeleteCampaign :exec
DELETE FROM campaigns WHERE id = $1;

-- name: CountCampaigns :one
SELECT COUNT(id)
FROM campaigns
WHERE changeset_ids ? $1;

-- name: GetCampaign :one
SELECT
  id,
  name,
  description,
  author_id,
  namespace_user_id,
  namespace_org_id,
  created_at,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
FROM campaigns
WHERE id = $1
LIMIT 1;

-- name: ListCampaigns :many
SELECT
  id,
  name,
  description,
  author_id,
  namespace_user_id,
  namespace_org_id,
  created_at,
  updated_at,
  changeset_ids,
  campaign_plan_id,
  closed_at
FROM campaigns
WHERE changeset_ids ? $1
ORDER BY id ASC
LIMIT $2;


-- name: CreateCampaignPlan :one
INSERT INTO campaign_plans (
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5)
RETURNING
  id,
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at;


-- name: UpdateCampaignPlan :many
UPDATE campaign_plans
SET (
  campaign_type,
  arguments,
  canceled_at,
  updated_at
) = ($1, $2, $3, $4)
WHERE id = $5
RETURNING
  id,
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at;

-- name: DeleteCampaignPlan :exec
DELETE FROM campaign_plans WHERE id = $1;

-- name: DeleteExpiredCampaignPlans :exec
DELETE FROM
  campaign_plans
WHERE
NOT EXISTS (
  SELECT 1
  FROM
  campaigns
  WHERE
  campaigns.campaign_plan_id = campaign_plans.id
)
AND
NOT EXISTS (
  SELECT id
  FROM
  campaign_jobs
  WHERE
  campaign_jobs.campaign_plan_id = campaign_plans.id
  AND
  (
    campaign_jobs.finished_at IS NULL
    OR
    campaign_jobs.finished_at > $1
  )
);


-- name: CountCampaignPlans :one
SELECT COUNT(id)
FROM campaign_plans;


-- name: GetCampaignPlan :one
SELECT
  id,
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at
FROM campaign_plans
WHERE id = $1
LIMIT 1;

-- name: GetCampaignPlanStatus :one
SELECT
  (SELECT canceled_at IS NOT NULL FROM campaign_plans WHERE campaign_plans.id = $1) AS canceled,
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE finished_at IS NULL) AS pending,
  COUNT(*) FILTER (WHERE finished_at IS NOT NULL AND (diff != '' OR error != '')) AS completed,
  array_agg(error) FILTER (WHERE error != '') AS errors
FROM campaign_jobs
WHERE campaign_plan_id = $2
LIMIT 1;

-- name: GetCampaignStatus :one
SELECT
  -- canceled is here so that this can be used with scanBackgroundProcessStatus
  false AS canceled,
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE finished_at IS NULL) AS pending,
  COUNT(*) FILTER (WHERE finished_at IS NOT NULL) AS completed,
  array_agg(error) FILTER (WHERE error != '') AS errors
FROM changeset_jobs
WHERE campaign_id = $1
LIMIT 1;

-- name: ListCampaignPlans :many
SELECT
  id,
  campaign_type,
  arguments,
  canceled_at,
  created_at,
  updated_at
FROM campaign_plans
WHERE id >= $1
ORDER BY id ASC
LIMIT $2;

-- name: CreateCampaignJob :one
INSERT INTO campaign_jobs (
  campaign_plan_id,
  repo_id,
  rev,
  base_ref,
  diff,
  description,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING
  id,
  campaign_plan_id,
  repo_id,
  rev,
  base_ref,
  diff,
  description,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at;

-- name: UpdateCampaignJob :many
UPDATE campaign_jobs
SET (
  campaign_plan_id,
  repo_id,
  rev,
  base_ref,
  diff,
  description,
  error,
  started_at,
  finished_at,
  updated_at
) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
WHERE id = $11
RETURNING
  id,
  campaign_plan_id,
  repo_id,
  rev,
  base_ref,
  diff,
  description,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at;


-- name: DeleteCampaignJob :exec
DELETE FROM campaign_jobs WHERE id = $1;


-- name: CountCampaignJobs :one
SELECT COUNT(id)
FROM campaign_jobs
WHERE campaign_plan_id = $1;

-- name: GetCampaignJob :one
SELECT
  id,
  campaign_plan_id,
  repo_id,
  rev,
  base_ref,
  diff,
  description,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
FROM campaign_jobs
WHERE id = $1
LIMIT 1;

-- name: ListCampaignJobs :many
SELECT
  id,
  campaign_plan_id,
  repo_id,
  rev,
  base_ref,
  diff,
  description,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
FROM campaign_jobs
WHERE campaign_plan_id = $1
ORDER BY id ASC;

-- name: CreateChangesetJob :one
INSERT INTO changeset_jobs (
  campaign_id,
  campaign_job_id,
  changeset_id,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING
  id,
  campaign_id,
  campaign_job_id,
  changeset_id,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at;

-- name: UpdateChangesetJob :many
UPDATE changeset_jobs
SET (
  campaign_id,
  campaign_job_id,
  changeset_id,
  error,
  started_at,
  finished_at,
  updated_at
) = ($1, $2, $3, $4, $5, $6, $7)
WHERE id = $8
RETURNING
  id,
  campaign_id,
  campaign_job_id,
  changeset_id,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at;

-- name: DeleteChangesetJob :exec
DELETE FROM changeset_jobs WHERE id = $1;


-- name: CountChangesetJobs :one
SELECT COUNT(id)
FROM changeset_jobs
WHERE campaign_id = $1;


-- name: GetChangesetJob :one
SELECT
  id,
  campaign_id,
  campaign_job_id,
  changeset_id,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
FROM changeset_jobs
WHERE id = $1
LIMIT 1;

-- name: ListChangesetJobs :many
SELECT
  id,
  campaign_id,
  campaign_job_id,
  changeset_id,
  error,
  started_at,
  finished_at,
  created_at,
  updated_at
FROM changeset_jobs
WHERE campaign_id = $1
ORDER BY id ASC;

-- name: ResetFailedChangesetJobs :exec
UPDATE changeset_jobs
SET
  error = '',
  started_at = NULL,
  finished_at = NULL
WHERE campaign_id = $1;
