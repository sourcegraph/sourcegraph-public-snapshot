package billing

import (
	"errors"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	stripe "github.com/stripe/stripe-go"
)

// ToSubscriptionItemsParams converts a value of GraphQL type ProductSubscriptionInput into a
// subscription item parameter for the billing system.
func ToSubscriptionItemsParams(input graphqlbackend.ProductSubscriptionInput) *stripe.SubscriptionItemsParams {
	return &stripe.SubscriptionItemsParams{
		Plan:     stripe.String(input.BillingPlanID),
		Quantity: stripe.Int64(int64(input.UserCount)),
	}
}

// GetSubscriptionItemIDToReplace returns the ID of the billing subscription item (used when
// updating the subscription or previewing an invoice to do so). It also performs a good set of
// sanity checks on the subscription that should be performed whenever the subscription is updated.
func GetSubscriptionItemIDToReplace(billingSub *stripe.Subscription, billingCustomerID string) (string, error) {
	if billingSub.Customer.ID != billingCustomerID {
		return "", errors.New("product subscription's billing customer does not match the provided account parameter")
	}
	if len(billingSub.Items.Data) != 1 {
		return "", fmt.Errorf("product subscription has unexpected number of invoice items (got %d, want 1)", len(billingSub.Items.Data))
	}
	return billingSub.Items.Data[0].ID, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_649(size int) error {
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
