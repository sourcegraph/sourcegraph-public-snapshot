package platform

// HTTP response header keys that platform apps can set. Other
// response headers from the app are cleared and ignored (unless
// X-Sourcegraph-Verbatim is true).
const (
	// HTTPHeaderVerbatim will make Sourcegraph forward the application
	// HTTP response directly to the user. This is useful for things
	// like static assets and API endpoints.
	HTTPHeaderVerbatim = "X-Sourcegraph-Verbatim"

	// HTTPHeaderTitle customizes an app page's <title>.
	HTTPHeaderTitle = "X-Sourcegraph-Title"
)
