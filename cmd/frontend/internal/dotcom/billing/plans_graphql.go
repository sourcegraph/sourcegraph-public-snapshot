package billing

import (
	"context"
	"fmt"
	"sort"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
)

// productPlan implements the GraphQL type ProductPlan.
type productPlan struct {
	billingID           string
	name                string
	title               string
	pricePerUserPerYear int32
}

func (r *productPlan) BillingID() string          { return r.billingID }
func (r *productPlan) Name() string               { return r.name }
func (r *productPlan) Title() string              { return r.title }
func (r *productPlan) FullProductName() string    { return "Sourcegraph " + r.title }
func (r *productPlan) PricePerUserPerYear() int32 { return r.pricePerUserPerYear }

// ToProductPlan returns a resolver for the GraphQL type ProductPlan from the given billing plan.
func ToProductPlan(plan *stripe.Plan) (graphqlbackend.ProductPlan, error) {
	// Sanity check.
	if plan.BillingScheme != stripe.PlanBillingSchemePerUnit {
		return nil, fmt.Errorf("unexpected billing scheme %q for plan %q", plan.BillingScheme, plan.ID)
	}
	if plan.TiersMode != "" {
		return nil, fmt.Errorf("unexpected tier mode %q for plan %q", plan.TiersMode, plan.ID)
	}
	if plan.Currency != stripe.CurrencyUSD {
		return nil, fmt.Errorf("unexpected currency %q for plan %q", plan.Currency, plan.ID)
	}
	if plan.Interval != stripe.PlanIntervalYear {
		return nil, fmt.Errorf("unexpected plan interval %q for plan %q", plan.Interval, plan.ID)
	}
	if plan.IntervalCount != 1 {
		return nil, fmt.Errorf("unexpected plan interval count %d for plan %q", plan.IntervalCount, plan.ID)
	}

	gqlPlan := &productPlan{
		billingID:           plan.ID,
		name:                plan.Nickname,
		title:               plan.Metadata["title"],
		pricePerUserPerYear: int32(plan.Amount),
	}
	var err error
	gqlPlan.title, err = getProductPlanTitle(plan)
	if err != nil {
		return nil, err
	}
	return gqlPlan, nil
}

// ProductPlans implements the GraphQL field Query.dotcom.productPlans.
func (BillingResolver) ProductPlans(ctx context.Context) ([]graphqlbackend.ProductPlan, error) {
	plans := plan.List(&stripe.PlanListParams{
		ListParams: stripe.ListParams{Context: ctx},
		Active:     stripe.Bool(true),
		Product:    stripe.String(stripeProductID),
	})
	var gqlPlans []graphqlbackend.ProductPlan
	for plans.Next() {
		gqlPlan, err := ToProductPlan(plans.Plan())
		if err != nil {
			return nil, err
		}
		gqlPlans = append(gqlPlans, gqlPlan)
	}
	if err := plans.Err(); err != nil {
		return nil, err
	}
	sort.Slice(gqlPlans, func(i, j int) bool {
		return gqlPlans[i].Name() < gqlPlans[j].Name()
	})
	return gqlPlans, nil
}
