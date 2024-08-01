package graphql

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/stretchr/testify/require"
)

func TestOmittingSyntacticField(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	observationCtx := observation.TestContextTB(t)

	service := policies.NewService(observationCtx, db, &mockUploadService{}, gitserver.NewMockClient())
	resolver := NewRootResolver(observationCtx, service, db.Repos(), lenientSiteadminchecker{})

	policyArgs :=
		func(syntacticIndexingEnabled *bool) resolvers.CodeIntelConfigurationPolicy {
			return resolvers.CodeIntelConfigurationPolicy{
				Name:                      "test",
				Type:                      "GIT_COMMIT",
				Pattern:                   "HEAD",
				RetentionEnabled:          false,
				RetainIntermediateCommits: true,
				RetentionDurationHours:    intptr(0),
				IndexingEnabled:           true,
				SyntacticIndexingEnabled:  syntacticIndexingEnabled,
				IndexCommitMaxAgeHours:    intptr(8064),
				IndexIntermediateCommits:  true,
			}
		}

	// By default the syntactic indexing is disabled if request doesn't send an explicit value
	created, err := resolver.CreateCodeIntelligenceConfigurationPolicy(ctx, &resolvers.CreateCodeIntelligenceConfigurationPolicyArgs{
		CodeIntelConfigurationPolicy: policyArgs(nil),
	})
	require.NoError(t, err)
	require.False(t, *created.SyntacticIndexingEnabled())

	// Explicitly setting syntactic indexing field to true
	_, err = resolver.UpdateCodeIntelligenceConfigurationPolicy(ctx, &resolvers.UpdateCodeIntelligenceConfigurationPolicyArgs{
		ID:                           created.ID(),
		CodeIntelConfigurationPolicy: policyArgs(boolptr(true)),
	})
	require.NoError(t, err)
	retrieved, err := resolver.ConfigurationPolicyByID(ctx, created.ID())
	require.NoError(t, err)
	require.True(t, *retrieved.SyntacticIndexingEnabled())

	// Omitting the syntactic indexing field during update should lead to no changes
	_, err = resolver.UpdateCodeIntelligenceConfigurationPolicy(ctx, &resolvers.UpdateCodeIntelligenceConfigurationPolicyArgs{
		ID:                           created.ID(),
		CodeIntelConfigurationPolicy: policyArgs(nil),
	})
	require.NoError(t, err)
	retrieved, err = resolver.ConfigurationPolicyByID(ctx, created.ID())
	require.NoError(t, err)
	require.True(t, *retrieved.SyntacticIndexingEnabled())

	// Setting syntactic indexing field explicitly propagates the value into the policy
	createdAnother, err := resolver.CreateCodeIntelligenceConfigurationPolicy(ctx, &resolvers.CreateCodeIntelligenceConfigurationPolicyArgs{
		CodeIntelConfigurationPolicy: policyArgs(boolptr(true)),
	})
	fmt.Println(createdAnother)
	require.NoError(t, err)
	require.NotEqual(t, created.ID(), createdAnother.ID())
	require.True(t, *createdAnother.SyntacticIndexingEnabled())
}

type lenientSiteadminchecker struct{}

// CheckCurrentUserIsSiteAdmin implements sharedresolvers.SiteAdminChecker.
func (l lenientSiteadminchecker) CheckCurrentUserIsSiteAdmin(ctx context.Context) error {
	return nil
}

var _ sharedresolvers.SiteAdminChecker = lenientSiteadminchecker{}

func intptr(i int32) *int32 {
	return &i
}

func boolptr(i bool) *bool {
	return &i
}

type mockUploadService struct{}

// GetCommitsVisibleToUpload implements policies.UploadService.
func (m *mockUploadService) GetCommitsVisibleToUpload(ctx context.Context, uploadID int, limit int, token *string) (_ []string, nextToken *string, err error) {
	panic("unimplemented")
}

var _ policies.UploadService = &mockUploadService{}
