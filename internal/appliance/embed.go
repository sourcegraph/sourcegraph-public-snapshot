package appliance

import "embed"

var (
	//go:embed ui/static
	staticFS embed.FS

	//go:embed ui/template
	templateFS embed.FS
)
