package schema

// FrontendCodyProConfig is the configuration data for Cody Pro that
// needs to be passed to the frontend. This needs to be kept in-sync
// with the data type on the frontend-side, in
// client/web/src/jscontext.ts.
type FrontendCodyProConfig struct {
	StripePublishableKey string `json:"stripePublishableKey"`
}

func MakeFrontendCodyProConfig(config *CodyProConfig) *FrontendCodyProConfig {
	return &FrontendCodyProConfig{
		StripePublishableKey: config.StripePublishableKey,
	}
}
