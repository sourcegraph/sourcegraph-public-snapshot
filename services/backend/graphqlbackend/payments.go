package graphqlbackend

import (
	"context"
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/stripe"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

type Plan interface {
	Name() string
	Seats() *int32
	Organization(ctx context.Context) (*org, error)
	Cost() int32
	RenewalDate() *int32
}

type planResolver struct {
	*stripe.Plan
}

func (p planResolver) Seats() *int32 {
	n := int32(p.Plan.Seats)
	return &n
}

func (p planResolver) Name() string {
	return "organization"
}

func (p planResolver) Organization(ctx context.Context) (*org, error) {
	// Fetch from GH.
	organization, _, err := github.OrgsFromContext(ctx).Get(p.OrgName)
	if err != nil {
		return nil, err
	}
	return &org{*organization}, nil
}

func (p planResolver) Cost() int32 {
	return int32(p.Plan.Cost)
}

func (p planResolver) RenewalDate() *int32 {
	n := int32(p.Plan.RenewalDate)
	return &n
}

type fakePlan struct {
	name string
}

func (f fakePlan) Seats() *int32 {
	return nil
}

func (f fakePlan) Name() string {
	return f.name
}

func (f fakePlan) Organization(ctx context.Context) (*org, error) {
	return nil, nil
}

func (f fakePlan) Cost() int32 {
	return 0
}
func (f fakePlan) RenewalDate() *int32 {
	return nil
}

func (r *currentUserResolver) PaymentPlan(ctx context.Context) (Plan, error) {
	// get stripe plan for user, else private.
	plan := stripe.GetPlan(ctx)
	if plan != nil {
		return planResolver{plan}, nil
	}

	private, err := AuthedPrivate(ctx)
	if err != nil {
		return nil, err
	}
	if private {
		return fakePlan{name: "private"}, nil
	}
	return fakePlan{name: "public"}, nil
}

func (*schemaResolver) CancelSubscription(ctx context.Context) (bool, error) {
	return true, stripe.CancelSubscription(ctx)
}

func AuthedPrivate(ctx context.Context) (bool, error) {
	appMeta, err := auth.GetAppMetadata(ctx)
	if err != nil {
		return false, err
	}
	ghScope, ok := appMeta["github_scope"].([]interface{})
	if !ok {
		return false, errors.New("unexpected type unmarshaling Auth0 metadata")
	}

	return listContains(ghScope, "repo"), nil
}

func listContains(l []interface{}, v interface{}) bool {
	for _, e := range l {
		if e == v {
			return true
		}
	}
	return false
}

func (*schemaResolver) UpdatePaymentSource(ctx context.Context, args *struct {
	TokenID string
}) (bool, error) {
	return true, stripe.SetTokenSourceForCustomer(ctx, args.TokenID)
}
