package billing

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/sourcegraph/enterprise/pkg/license"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/plan"
)

// InfoForProductPlan returns the license key tags and min quantity that should be used for the
// given product plan.
//
// License key tags indicate which product variant (e.g., Enterprise vs. Enterprise Starter), so
// they are stored on the billing system in the metadata of the product plans.
func InfoForProductPlan(ctx context.Context, planID string) (licenseTags []string, minQuantity *int32, err error) {
	params := &stripe.PlanParams{Params: stripe.Params{Context: ctx}}
	params.AddExpand("product")
	plan, err := plan.Get(planID, params)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, fmt.Errorf("unable to determine license tags for plan %q (nickname %q)", planID, plan.Nickname)
	}
	return tags, ProductPlanMinQuantity(plan), nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_646(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
