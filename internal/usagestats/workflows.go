package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type workflows struct {
	TotalWorkflows int32

	UniqueUserWorkflowOwners int32
	UniqueOrgWorkflowOwners  int32
	UserOwnedWorkflows       int32
	OrgOwnedWorkflows        int32

	WorkflowsCreatedLast24h int32
	WorkflowsUpdatedLast24h int32
}

func GetWorkflows(ctx context.Context, db database.DB) (*workflows, error) {
	const q = `
	SELECT
	(SELECT COUNT(*) FROM workflows) AS totalWorkflows,
	(SELECT COUNT(DISTINCT owner_user_id) FROM workflows) AS uniqueUserWorkflowOwners,
	(SELECT COUNT(DISTINCT owner_org_id) FROM workflows) AS uniqueOrgWorkflowOwners,
	(SELECT COUNT(*) FROM workflows WHERE owner_user_id IS NOT NULL) AS userOwnedWorkflows,
	(SELECT COUNT(*) FROM workflows WHERE owner_org_id IS NOT NULL) AS orgOwnedWorkflows,
	(SELECT COUNT(*) FROM workflows WHERE created_at > NOW() - INTERVAL '24 hours') AS workflowsCreatedLast24h,
	(SELECT COUNT(*) FROM workflows WHERE updated_at > NOW() - INTERVAL '24 hours') AS workflowsUpdatedLast24h
	`
	var v workflows
	if err := db.QueryRowContext(ctx, q).Scan(
		&v.TotalWorkflows,
		&v.UniqueUserWorkflowOwners,
		&v.UniqueOrgWorkflowOwners,
		&v.UserOwnedWorkflows,
		&v.OrgOwnedWorkflows,
		&v.WorkflowsCreatedLast24h,
		&v.WorkflowsUpdatedLast24h,
	); err != nil {
		return nil, err
	}

	return &v, nil
}
