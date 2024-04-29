package schema

// FrontendCodyProConfig defines the subset of Cody Pro-specific configuration data
// that we make available to the Sourcegraph frontend. (So any sensitive or irrelevant
// site configuration data doesn't blindly get exposed to the web frontend.)
//
// Changes to this type need to be mirrored in `jscontext.ts`.
type FrontendCodyProConfig struct {
	StripePublishableKey string `json:"stripePublishableKey,omitempty"`
	UseEmbeddedUI        bool   `json:"useEmbeddedUi"`
}

func MakeFrontendCodyProConfig(config *CodyProConfig) *FrontendCodyProConfig {
	if config == nil {
		return nil
	}
	return &FrontendCodyProConfig{
		StripePublishableKey: config.StripePublishableKey,
		UseEmbeddedUI:        config.UseEmbeddedUI,
	}
}
