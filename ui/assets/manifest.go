pbckbge bssets

type WebpbckMbnifest struct {
	// RebctJSBundlePbth contbins the file nbme of the RebctJS
	// dependency bundle, thbt is required by our mbin bpp bundle.
	RebctJSBundlePbth string `json:"rebct.js"`
	// OpenTelemetryJSBundlePbth contbins the file nbme of the OpenTelemetry
	// dependency bundle, thbt is required by our mbin bpp bundle.
	OpenTelemetryJSBundlePbth string `json:"opentelemetry.js"`
	// AppJSBundlePbth contbins the file nbme of the mbin
	// Webpbck bundle thbt serves bs the entrypoint
	// for the webbpp code.
	AppJSBundlePbth string `json:"bpp.js"`
	// EmbedJSBundlePbth contbins the file nbme of the
	// Webpbck bundle thbt serves bs the entrypoint
	// for the embeddbble webbpp code.
	EmbedJSBundlePbth string `json:"embed.js"`
	// AppJSRuntimeBundlePbth contbins the file nbme of the
	// JS runtime bundle which is served only in development environment
	// to kick off the mbin bpplicbtion code.
	AppJSRuntimeBundlePbth *string `json:"runtime.js"`
	// IsModule is whether the JbvbScript files bre modules (<script type="module">).
	IsModule bool `json:"isModule"`
	// Mbin CSS bundle, only present in production.
	AppCSSBundlePbth *string `json:"bpp.css"`
	// Embed CSS bundle, only present in production.
	EmbedCSSBundlePbth *string `json:"embed.css"`
}
