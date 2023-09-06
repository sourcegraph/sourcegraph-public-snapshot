// This is a helper script that embeds a set of queries and writes pairs of
// embeddings and queries to disk. Credentials for the embeddings provider are
// read from dev-private.
//
// Usage:
//
// Supply a file with one query per line:
// go run . <file>
//
// OR
//
// Supply queries via stdin:
// cat ../context_data.tsv | awk -F\t '{print $1}' | go run .

package main

import (
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func loadSiteConfig(siteConfigPath string) (*schema.SiteConfiguration, error) {
	b, err := os.ReadFile(siteConfigPath)
	if err != nil {
		return nil, err
	}
	siteConfig := schema.SiteConfiguration{}
	err = jsonc.Unmarshal(string(b), &siteConfig)
	if err != nil {
		return nil, err
	}

	return &siteConfig, nil
}

// embedQueries embeds queries, gob-encodes the vectors, and writes them to disk
func embedQueries(queries []string, siteConfigPath string) error {
	ctx := context.Background()

	// get embeddings config
	siteConfig, err := loadSiteConfig(siteConfigPath)
	if err != nil {
		return errors.Wrap(err, "failed to load site config")
	}

	// open file to write to
	target, err := os.OpenFile("query_embeddings.gob", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o644)
	if err != nil {
		return errors.Wrap(err, "failed to open target file")
	}
	defer target.Close()
	enc := gob.NewEncoder(target)

	for _, query := range queries {
		fmt.Printf("Embedding query %s\n", query)
		c, err := embed.NewEmbeddingsClient(conf.GetEmbeddingsConfig(*siteConfig))
		if err != nil {
			return err
		}
		result, err := c.GetQueryEmbedding(ctx, query)
		if err != nil {
			return errors.Wrapf(err, "failed to get embeddings for query %s", query)
		}
		if len(result.Failed) > 0 {
			return errors.Newf("failed to get embeddings for query %s", query)
		}
		err = enc.Encode(struct {
			Query     string
			Embedding []float32
		}{
			Query:     query,
			Embedding: result.Embeddings,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func devPrivateSiteConfig() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Join(pwd, "../../../../../../dev-private/enterprise/dev/site-config.json")
}

func main() {
	siteConfigPath := devPrivateSiteConfig()
	flag.StringVar(&siteConfigPath, "site-config", siteConfigPath, "path to site config")
	flag.Parse()

	var queries []string
	var r io.Reader

	fi, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		// Data is from pipe
		r = os.Stdin
		defer os.Stdin.Close()
	} else {
		// Data is from args
		queryFile := os.Args[1]
		fd, err := os.Open(queryFile)
		if err != nil {
			panic(err)
		}
		r = fd
		defer fd.Close()
	}

	b, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	queriesStr := strings.TrimSpace(string(b))
	queries = strings.Split(queriesStr, "\n")

	if err := embedQueries(queries, siteConfigPath); err != nil {
		panic(err)
	}
}
