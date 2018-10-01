package billing

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/enterprise/pkg/license"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
)

// getProductPlanTitle returns the title for the plan if any title is found (or can be guessed). If
// there is no title, it returns an error.
func getProductPlanTitle(plan *stripe.Plan) (string, error) {
	var title string
	switch {
	case plan.Metadata["title"] != "":
		title = plan.Metadata["title"]
	case isMaybeEnterpriseStarter(plan.Nickname):
		title = "Enterprise Starter"
	case isMaybeEnterprise(plan.Nickname):
		title = "Enterprise"
	}
	if title == "" {
		return "", fmt.Errorf("unexpected empty title for plan %q", plan.ID)
	}
	return title, nil
}

func isMaybeEnterpriseStarter(planNickname string) bool {
	return strings.Contains(planNickname, "enterprise-starter")
}

func isMaybeEnterprise(planNickname string) bool {
	return strings.Contains(planNickname, "enterprise") && !isMaybeEnterpriseStarter(planNickname)
}

// LicenseTagsForProductPlan returns the license key tags that should be used for the given product
// plan. License key tags indicate which product variant (e.g., Enterprise vs. Enterprise Starter),
// so they are stored on the billing system in the metadata of the product plans.
func LicenseTagsForProductPlan(ctx context.Context, planID string) ([]string, error) {
	plan, err := plan.Get(planID, &stripe.PlanParams{Params: stripe.Params{Context: ctx}})
	if err != nil {
		return nil, err
	}

	var tags []string
	switch {
	case plan.Metadata["licenseTags"] != "":
		tags = license.ParseTagsInput(plan.Metadata["licenseTags"])
	case isMaybeEnterpriseStarter(plan.Nickname):
		tags = licensing.EnterpriseStarterTags
	case isMaybeEnterprise(plan.Nickname):
		tags = licensing.EnterpriseTags
	default:
		return nil, fmt.Errorf("unable to determine license tags for plan %q (nickname %q)", planID, plan.Nickname)
	}
	return tags, nil
}
