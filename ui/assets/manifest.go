package assets

type WebBuildManifest struct {
	MainJSBundlePath   string `json:"main.js"`
	MainCSSBundlePath  string `json:"main.css"`
	EmbedJSBundlePath  string `json:"embed.js"`
	EmbedCSSBundlePath string `json:"embed.css"`
}
