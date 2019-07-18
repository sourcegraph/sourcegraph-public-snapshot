package rules

import (
	"context"
	"path"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/projects"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlRule implements the GraphQL type Rule.
type gqlRule struct{ db *dbRule }

// ruleByID looks up and returns the Rule with the given GraphQL ID. If no such Rule exists, it
// returns a non-nil error.
func ruleByID(ctx context.Context, id graphql.ID) (*gqlRule, error) {
	dbID, err := unmarshalRuleID(id)
	if err != nil {
		return nil, err
	}
	return ruleByDBID(ctx, dbID)
}

func (GraphQLResolver) RuleByID(ctx context.Context, id graphql.ID) (graphqlbackend.Rule, error) {
	return ruleByID(ctx, id)
}

// ruleByDBID looks up and returns the Rule with the given database ID. If no such Rule exists,
// it returns a non-nil error.
func ruleByDBID(ctx context.Context, dbID int64) (*gqlRule, error) {
	v, err := dbRules{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return &gqlRule{db: v}, nil
}

func (v *gqlRule) ID() graphql.ID {
	return marshalRuleID(v.db.ID)
}

func marshalRuleID(id int64) graphql.ID {
	return relay.MarshalID("Rule", id)
}

func unmarshalRuleID(id graphql.ID) (dbID int64, err error) {
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

func (v *gqlRule) Project(ctx context.Context) (graphqlbackend.Project, error) {
	return graphqlbackend.ProjectByDBID(ctx, v.db.ProjectID)
}

func (v *gqlRule) Name() string { return v.db.Name }

func (v *gqlRule) Description() *string { return v.db.Description }

func (v *gqlRule) Settings() string { return v.db.Settings }

func (v *gqlRule) URL() string {
	return path.Join(projects.URLToProject(v.db.ProjectID), "rules", string(v.ID()))
}
