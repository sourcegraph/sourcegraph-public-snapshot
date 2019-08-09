package billing

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
)

func init() {
	graphqlbackend.UserURLForSiteAdminBilling = func(ctx context.Context, userID int32) (*string, error) {
		// 🚨 SECURITY: Only site admins may view the billing URL, because it may contain sensitive
		// data or identifiers.
		if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
			return nil, err
		}
		custID, err := dbBilling{}.getUserBillingCustomerID(ctx, nil, userID)
		if err != nil {
			return nil, err
		}
		if custID != nil {
			u := CustomerURL(*custID)
			return &u, nil
		}
		return nil, nil
	}
}

func (BillingResolver) SetUserBilling(ctx context.Context, args *graphqlbackend.SetUserBillingArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may set a user's billing info.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := graphqlbackend.UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	// Ensure the billing customer ID refers to a valid customer in the billing system.
	if args.BillingCustomerID != nil {
		if _, err := customer.Get(*args.BillingCustomerID, &stripe.CustomerParams{Params: stripe.Params{Context: ctx}}); err != nil {
			return nil, err
		}
	}

	if err := (dbBilling{}).setUserBillingCustomerID(ctx, nil, userID, args.BillingCustomerID); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_639(size int) error {
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
