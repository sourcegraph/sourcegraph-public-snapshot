package assets

type WebpackManifest struct {
	// ReactJSBundlePath contains the file name of the ReactJS
	// dependency bundle, that is required by our main app bundle.
	ReactJSBundlePath string `json:"react.js"`
	// AppJSBundlePath contains the file name of the main
	// Webpack bundle that serves as the entrypoint
	// for the webapp code.
	AppJSBundlePath string `json:"app.js"`
	// EmbedJSBundlePath contains the file name of the
	// Webpack bundle that serves as the entrypoint
	// for the embeddable webapp code.
	EmbedJSBundlePath string `json:"embed.js"`
	// AppJSRuntimeBundlePath contains the file name of the
	// JS runtime bundle which is served only in development environment
	// to kick off the main application code.
	AppJSRuntimeBundlePath *string `json:"runtime.js"`
	// IsModule is whether the JavaScript files are modules (<script type="module">).
	IsModule bool `json:"isModule"`
	// Main CSS bundle, only present in production.
	AppCSSBundlePath *string `json:"app.css"`
	// Embed CSS bundle, only present in production.
	EmbedCSSBundlePath *string `json:"embed.css"`
}
