package rules

import (
	"context"
	"path"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlRule implements the GraphQL type Rule.
type gqlRule struct{ db *DBRule }

// ruleByID looks up and returns the Rule with the given GraphQL ID. If no such Rule exists, it
// returns a non-nil error.
func ruleByID(ctx context.Context, id graphql.ID) (*gqlRule, error) {
	dbID, err := graphqlbackend.UnmarshalRuleID(id)
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
	v, err := DBRules{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return &gqlRule{db: v}, nil
}

func (v *gqlRule) ID() graphql.ID {
	return graphqlbackend.MarshalRuleID(v.db.ID)
}

func (v *gqlRule) Container(ctx context.Context) (*graphqlbackend.ToRuleContainer, error) {
	return GraphQLResolver{}.RuleContainerByID(ctx, v.db.Container.graphqlID())
}

func (v *gqlRule) Name() string { return v.db.Name }

func (v *gqlRule) Description() *string { return v.db.Description }

func (v *gqlRule) Definition() graphqlbackend.JSONC { return graphqlbackend.JSONC(v.db.Definition) }

func (v *gqlRule) CreatedAt() graphqlbackend.DateTime { return graphqlbackend.DateTime{v.db.CreatedAt} }

func (v *gqlRule) UpdatedAt() graphqlbackend.DateTime { return graphqlbackend.DateTime{v.db.UpdatedAt} }

func (v *gqlRule) URL(ctx context.Context) (string, error) {
	toContainer, err := v.Container(ctx)
	if err != nil {
		return "", err
	}
	var containerURL string
	if u, ok := toContainer.RuleContainer().(interface {
		URL(context.Context) (string, error)
	}); ok {
		var err error
		containerURL, err = u.URL(ctx)
		if err != nil {
			return "", err
		}
	} else if u, ok := toContainer.RuleContainer().(interface{ URL() string }); ok {
		containerURL = u.URL()
	} else {
		return "", errors.New("unable to get rule container URL")
	}
	return path.Join(containerURL, "rules", string(v.ID())), nil
}

func (v *gqlRule) ViewerCanUpdate(ctx context.Context) (bool, error) {
	toContainer, err := v.Container(ctx)
	if err != nil {
		return false, err
	}
	if u, ok := toContainer.RuleContainer().(graphqlbackend.Updatable); ok {
		return u.ViewerCanUpdate(ctx)
	}
	return false, errors.New("unable to determine viewerCanUpdate")
}
