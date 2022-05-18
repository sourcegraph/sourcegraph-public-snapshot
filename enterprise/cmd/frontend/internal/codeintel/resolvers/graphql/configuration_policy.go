package graphql

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type configurationPolicyResolver struct {
	db                  database.DB
	configurationPolicy store.ConfigurationPolicy
	errTracer           *observation.ErrCollector
}

func NewConfigurationPolicyResolver(db database.DB, configurationPolicy store.ConfigurationPolicy, errTracer *observation.ErrCollector) gql.CodeIntelligenceConfigurationPolicyResolver {
	return &configurationPolicyResolver{
		db:                  db,
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

func (r *configurationPolicyResolver) Repository(ctx context.Context) (_ *gql.RepositoryResolver, err error) {
	if r.configurationPolicy.RepositoryID == nil {
		return nil, nil
	}

	defer r.errTracer.Collect(&err,
		log.String("configurationPolicyResolver.field", "repository"),
		log.Int("configurationPolicyID", r.configurationPolicy.ID),
		log.Int("repoID", *r.configurationPolicy.RepositoryID),
	)

	repo, err := backend.NewRepos(r.db).Get(ctx, api.RepoID(*r.configurationPolicy.RepositoryID))
	if err != nil {
		return nil, err
	}

	return gql.NewRepositoryResolver(r.db, repo), nil
}

func (r *configurationPolicyResolver) RepositoryPatterns() *[]string {
	return r.configurationPolicy.RepositoryPatterns
}

func (r *configurationPolicyResolver) Type() (_ gql.GitObjectType, err error) {
	defer r.errTracer.Collect(&err,
		log.String("configurationPolicyResolver.field", "type"),
		log.Int("configurationPolicyID", r.configurationPolicy.ID),
		log.String("policyType", string(r.configurationPolicy.Type)),
	)

	switch r.configurationPolicy.Type {
	case store.GitObjectTypeCommit:
		return gql.GitObjectTypeCommit, nil
	case store.GitObjectTypeTag:
		return gql.GitObjectTypeTag, nil
	case store.GitObjectTypeTree:
		return gql.GitObjectTypeTree, nil
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
