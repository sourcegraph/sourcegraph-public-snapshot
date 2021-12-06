// package migrations contains embedded migrate sources for our DB migrations.
package migrations

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed frontend/*.sql codeintel/*.sql codeinsights/*.sql
var QueryDefinitions embed.FS

var (
	CodeInsights = mustSub("codeinsights")
	CodeIntel    = mustSub("codeintel")
	Frontend     = mustSub("frontend")
)

func mustSub(dir string) fs.FS {
	f, err := fs.Sub(QueryDefinitions, dir)
	if err != nil {
		log.Fatalf("could not create DB migration fs %s: %v", dir, err)
	}
	return f
}
