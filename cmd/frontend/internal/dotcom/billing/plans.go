package billing

import (
	"context"
	"fmt"

	"github.com/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/enterprise/pkg/license"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
)

// LicenseTagsForProductPlan returns the license key tags that should be used for the given product
// plan. License key tags indicate which product variant (e.g., Enterprise vs. Enterprise Starter),
// so they are stored on the billing system in the metadata of the product plans.
func LicenseTagsForProductPlan(ctx context.Context, planID string) ([]string, error) {
	params := &stripe.PlanParams{Params: stripe.Params{Context: ctx}}
	params.AddExpand("product")
	plan, err := plan.Get(planID, params)
	if err != nil {
		return nil, err
	}

	var tags []string
	switch {
	case plan.Product.Metadata["licenseTags"] != "":
		tags = license.ParseTagsInput(plan.Product.Metadata["licenseTags"])
	case plan.Product.Name == "Enterprise Starter":
		tags = licensing.EnterpriseStarterTags
	case plan.Product.Name == "Enterprise":
		tags = licensing.EnterpriseTags
	default:
		return nil, fmt.Errorf("unable to determine license tags for plan %q (nickname %q)", planID, plan.Nickname)
	}
	return tags, nil
}
