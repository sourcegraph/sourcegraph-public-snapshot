package main

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
)

func main() {
	gen := &generator{
		outDir:        "sitemap/",
		queryDatabase: "sitemap_cache/sitemap_query.db",
	}
	if err := gen.generate(); err != nil {
		log15.Error("failed to generate", err)
		os.Exit(-1)
	}
	log15.Info("generated sitemap", "out", gen.outDir)
}

type generator struct {
	outDir        string
	queryDatabase string
}

// generate generates the sitemap files to the specified directory.
func (g *generator) generate() error {
	if err := os.MkdirAll(g.outDir, 0700); err != nil {
		return errors.Wrap(err, "MkdirAll")
	}
	if err := os.MkdirAll(filepath.Dir(g.queryDatabase), 0700); err != nil {
		return errors.Wrap(err, "MkdirAll")
	}

	// The query database caches our GraphQL queries across multiple runs, as well as allows us to
	// update the sitemap to include new repositories / pages without re-querying everything which
	// would be very expensive. It's a simple on-disk key-vaue store (bbolt).
	db, err := openQueryDatabase(g.queryDatabase)
	if err != nil {
		return errors.Wrap(err, "openQueryDatabase")
	}
	defer db.close()

	return nil
}
