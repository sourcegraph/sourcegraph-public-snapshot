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
	billingPlanID       string
	name                string
	pricePerUserPerYear int32
}

func (r *productPlan) BillingPlanID() string      { return r.billingPlanID }
func (r *productPlan) Name() string               { return r.name }
func (r *productPlan) NameWithBrand() string      { return "Sourcegraph " + r.name }
func (r *productPlan) PricePerUserPerYear() int32 { return r.pricePerUserPerYear }

// ToProductPlan returns a resolver for the GraphQL type ProductPlan from the given billing plan.
func ToProductPlan(plan *stripe.Plan) (graphqlbackend.ProductPlan, error) {
	// Sanity check.
	if plan.Product.Name == "" {
		return nil, fmt.Errorf("unexpected empty product name for plan %q", plan.ID)
	}
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
	return &productPlan{
		billingPlanID:       plan.ID,
		name:                plan.Product.Name,
		pricePerUserPerYear: int32(plan.Amount),
	}, nil
}

// ProductPlans implements the GraphQL field Query.dotcom.productPlans.
func (BillingResolver) ProductPlans(ctx context.Context) ([]graphqlbackend.ProductPlan, error) {
	params := &stripe.PlanListParams{
		ListParams: stripe.ListParams{Context: ctx},
		Active:     stripe.Bool(true),
	}
	params.AddExpand("data.product")
	plans := plan.List(params)
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

	// Sort cheapest first (a reasonable assumption).
	sort.Slice(gqlPlans, func(i, j int) bool {
		return gqlPlans[i].PricePerUserPerYear() < gqlPlans[j].PricePerUserPerYear()
	})

	return gqlPlans, nil
}
