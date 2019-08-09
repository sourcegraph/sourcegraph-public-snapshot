package billing

import (
	"log"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/app"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	stripe "github.com/stripe/stripe-go"
)

var (
	stripeSecretKey      = env.Get("STRIPE_SECRET_KEY", "", "billing: Stripe API secret key")
	stripePublishableKey = env.Get("STRIPE_PUBLISHABLE_KEY", "", "billing: Stripe API publishable key")
)

func init() {
	// Sanity-check the Stripe keys (to help prevent mistakes where they got switched and the secret
	// key is published).
	if stripeSecretKey != "" && !strings.HasPrefix(stripeSecretKey, "sk_") {
		log.Fatal(`Invalid STRIPE_SECRET_KEY (must begin with "sk_").`)
	}
	if stripePublishableKey != "" && !strings.HasPrefix(stripePublishableKey, "pk_") {
		log.Fatal(`Invalid STRIPE_PUBLISHABLE_KEY (must begin with "pk_").`)
	}
	if (stripeSecretKey != "") != (stripePublishableKey != "") {
		log.Fatalf("Either zero or both of STRIPE_SECRET_KEY (set=%v) and STRIPE_PUBLISHABLE_KEY (set=%v) must be set.", stripeSecretKey != "", stripePublishableKey != "")
	}

	stripe.Key = stripeSecretKey
	app.SetBillingPublishableKey(stripePublishableKey)

	if env.InsecureDev && stripe.Key != "" && !isTest() {
		log.Fatal("Refusing to use Stripe live (non-test) API key in Sourcegraph dev mode. (Sourcegraphers: Update your local dev-private checkout if you're getting this unexpectedly.)")
	}
}

func isTest() bool {
	return strings.Contains(stripe.Key, "_test_")
}

func baseURL() string {
	u := "https://dashboard.stripe.com"
	if isTest() {
		u += "/test"
	}
	return u
}

// CustomerURL returns the URL to the customer with the given ID on the billing system.
func CustomerURL(id string) string {
	return baseURL() + "/customers/" + id
}

// SubscriptionURL returns the URL to the subscription with the given ID on the billing system.
func SubscriptionURL(id string) string {
	return baseURL() + "/subscriptions/" + id
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_648(size int) error {
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
