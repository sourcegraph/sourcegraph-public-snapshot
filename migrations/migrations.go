// package migrations contains embedded migrate sources for our DB migrations.
package migrations

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed codeinsights/* codeintel/* frontend/*
var content embed.FS

var (
	CodeInsights = mustSub("codeinsights")
	CodeIntel    = mustSub("codeintel")
	Frontend     = mustSub("frontend")
)

func mustSub(dir string) fs.FS {
	f, err := fs.Sub(content, dir)
	if err != nil {
		log.Fatalf("could not create DB migration fs %s: %v", dir, err)
	}
	return f
}
