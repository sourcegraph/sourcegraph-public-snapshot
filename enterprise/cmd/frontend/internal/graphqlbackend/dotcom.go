// Package graphqlbackend injects enterprise GraphQL resolvers into our main
// graphqlbackend package. It does this as a side-effect of being imported.
package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/billing"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
)

func init() {
	// Contribute the GraphQL types DotcomMutation and DotcomQuery.
	graphqlbackend.Dotcom = dotcomResolver{}
}

// dotcomResolver implements the GraphQL types DotcomMutation and DotcomQuery.
type dotcomResolver struct {
	productsubscription.ProductSubscriptionLicensingResolver
	billing.BillingResolver
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_662(size int) error {
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
