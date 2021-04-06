package assets

type WebpackManifest struct {
	// AppJSBundlePath contains the file name of the main
	// Webpack bundle that serves as the entrypoint
	// for the webapp code.
	AppJSBundlePath string `json:"app.js"`
	// Main CSS bundle, only present in production.
	AppCSSBundlePath *string `json:"app.css"`
}
