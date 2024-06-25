package appliance

import (
	"embed"
	"io/fs"
)

var (
	//go:embed web/static
	staticFiles embed.FS
	staticFS, _ = fs.Sub(staticFiles, "web/static")

	//go:embed web/template
	templateFS embed.FS
)
