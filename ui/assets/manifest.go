package assets

// WebBuildManifest describes the web build and is produced by Vite. Keep it in sync with `interface
// WebBuildManifest`.
type WebBuildManifest struct {
	// URL is the base URL for asset paths.
	URL string `json:"url,omitempty"`

	// Assets is a map of entrypoint (such as "src/enterprise/main.tsx") to its JavaScript and CSS assets.
	Assets struct {
		Main      *entryAssets `json:"src/enterprise/main"`
		EmbedMain *entryAssets `json:"src/enterprise/embed/embedMain"`
		AppMain   *entryAssets `json:"src/enterprise/app/main"`
	} `json:"assets"`

	// DevInjectHTML contains additional HTML <script> tags to inject in dev mode.
	DevInjectHTML string `json:"devInjectHTML,omitempty"`
}

type entryAssets struct {
	JS  string `json:"js,omitempty"`
	CSS string `json:"css,omitempty"`
}
