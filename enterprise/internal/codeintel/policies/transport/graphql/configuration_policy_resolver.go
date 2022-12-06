package graphql

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type configurationPolicyResolver struct {
	logger              sglog.Logger
	policySvc           *policies.Service
	configurationPolicy types.ConfigurationPolicy
	errTracer           *observation.ErrCollector
}

func NewConfigurationPolicyResolver(policySvc *policies.Service, configurationPolicy types.ConfigurationPolicy, errTracer *observation.ErrCollector) resolverstubs.CodeIntelligenceConfigurationPolicyResolver {
	return &configurationPolicyResolver{
		policySvc:           policySvc,
		logger:              sglog.Scoped("configurationPolicyResolver", ""),
		configurationPolicy: configurationPolicy,
		errTracer:           errTracer,
	}
}

func (r *configurationPolicyResolver) ID() graphql.ID {
	return marshalConfigurationPolicyGQLID(int64(r.configurationPolicy.ID))
}

func (r *configurationPolicyResolver) Name() string {
	return r.configurationPolicy.Name
}

func (r *configurationPolicyResolver) Repository(ctx context.Context) (_ resolverstubs.RepositoryResolver, err error) {
	if r.configurationPolicy.RepositoryID == nil {
		return nil, nil
	}

	defer r.errTracer.Collect(&err,
		log.String("configurationPolicyResolver.field", "repository"),
		log.Int("configurationPolicyID", r.configurationPolicy.ID),
		log.Int("repoID", *r.configurationPolicy.RepositoryID),
	)

	db := r.policySvc.GetUnsafeDB()
	repo, err := backend.NewRepos(r.logger, db, gitserver.NewClient(db)).Get(ctx, api.RepoID(*r.configurationPolicy.RepositoryID))
	if err != nil {
		return nil, err
	}

	return sharedresolvers.NewRepositoryResolver(r.policySvc.GetUnsafeDB(), repo), nil
}

func (r *configurationPolicyResolver) RepositoryPatterns() *[]string {
	return r.configurationPolicy.RepositoryPatterns
}

func (r *configurationPolicyResolver) Type() (_ resolverstubs.GitObjectType, err error) {
	defer r.errTracer.Collect(&err,
		log.String("configurationPolicyResolver.field", "type"),
		log.Int("configurationPolicyID", r.configurationPolicy.ID),
		log.String("policyType", string(r.configurationPolicy.Type)),
	)

	switch r.configurationPolicy.Type {
	case types.GitObjectTypeCommit:
		return resolverstubs.GitObjectTypeCommit, nil
	case types.GitObjectTypeTag:
		return resolverstubs.GitObjectTypeTag, nil
	case types.GitObjectTypeTree:
		return resolverstubs.GitObjectTypeTree, nil
	default:
		return "", errors.Errorf("unknown git object type %s", r.configurationPolicy.Type)
	}
}

func (r *configurationPolicyResolver) Pattern() string {
	return r.configurationPolicy.Pattern
}

func (r *configurationPolicyResolver) Protected() bool {
	return r.configurationPolicy.Protected
}

func (r *configurationPolicyResolver) RetentionEnabled() bool {
	return r.configurationPolicy.RetentionEnabled
}

func (r *configurationPolicyResolver) RetentionDurationHours() *int32 {
	return toHours(r.configurationPolicy.RetentionDuration)
}

func (r *configurationPolicyResolver) RetainIntermediateCommits() bool {
	return r.configurationPolicy.RetainIntermediateCommits
}

func (r *configurationPolicyResolver) IndexingEnabled() bool {
	return r.configurationPolicy.IndexingEnabled
}

func (r *configurationPolicyResolver) IndexCommitMaxAgeHours() *int32 {
	return toHours(r.configurationPolicy.IndexCommitMaxAge)
}

func (r *configurationPolicyResolver) IndexIntermediateCommits() bool {
	return r.configurationPolicy.IndexIntermediateCommits
}

func toHours(duration *time.Duration) *int32 {
	if duration == nil {
		return nil
	}

	v := int32(*duration / time.Hour)
	return &v
}
