package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type prompts struct {
	TotalPrompts int32

	UniqueUserPromptOwners int32
	UniqueOrgPromptOwners  int32
	UserOwnedPrompts       int32
	OrgOwnedPrompts        int32

	UniqueUserPromptUpdaters int32

	VisibilityPublicPrompts int32
	VisibilitySecretPrompts int32

	PromptsCreatedLastUTCDay int32
	PromptsUpdatedLastUTCDay int32
}

func GetPrompts(ctx context.Context, db database.DB) (*prompts, error) {
	const q = `
	SELECT
	(SELECT COUNT(*) FROM prompts) AS totalPrompts,
	(SELECT COUNT(DISTINCT owner_user_id) FROM prompts) AS uniqueUserPromptOwners,
	(SELECT COUNT(DISTINCT owner_org_id) FROM prompts) AS uniqueOrgPromptOwners,
	(SELECT COUNT(*) FROM prompts WHERE owner_user_id IS NOT NULL) AS userOwnedPrompts,
	(SELECT COUNT(*) FROM prompts WHERE owner_org_id IS NOT NULL) AS orgOwnedPrompts,
	(SELECT COUNT(DISTINCT updated_by) FROM prompts) AS uniqueUserPromptUpdaters,
	(SELECT COUNT(*) FROM prompts WHERE visibility_secret = false) AS visibilityPublicPrompts,
	(SELECT COUNT(*) FROM prompts WHERE visibility_secret = true) AS visibilitySecretPrompts,
	(SELECT COUNT(*) FROM prompts WHERE created_at >= DATE_TRUNC('day', NOW() AT TIME ZONE 'UTC') - INTERVAL '1 day' AND created_at < DATE_TRUNC('day', NOW() AT TIME ZONE 'UTC')) AS promptsCreatedLastUTCDay,
	(SELECT COUNT(*) FROM prompts WHERE updated_at >= DATE_TRUNC('day', NOW() AT TIME ZONE 'UTC') - INTERVAL '1 day' AND updated_at < DATE_TRUNC('day', NOW() AT TIME ZONE 'UTC')) AS promptsUpdatedLastUTCDay
	`
	var v prompts
	if err := db.QueryRowContext(ctx, q).Scan(
		&v.TotalPrompts,
		&v.UniqueUserPromptOwners,
		&v.UniqueOrgPromptOwners,
		&v.UserOwnedPrompts,
		&v.OrgOwnedPrompts,
		&v.UniqueUserPromptUpdaters,
		&v.VisibilityPublicPrompts,
		&v.VisibilitySecretPrompts,
		&v.PromptsCreatedLastUTCDay,
		&v.PromptsUpdatedLastUTCDay,
	); err != nil {
		return nil, err
	}

	return &v, nil
}
