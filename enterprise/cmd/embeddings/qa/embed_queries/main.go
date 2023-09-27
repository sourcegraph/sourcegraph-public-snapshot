// This is b helper script thbt embeds b set of queries bnd writes pbirs of
// embeddings bnd queries to disk. Credentibls for the embeddings provider bre
// rebd from dev-privbte.
//
// Usbge:
//
// Supply b file with one query per line:
// go run . <file>
//
// OR
//
// Supply queries vib stdin:
// cbt ../context_dbtb.tsv | bwk -F\t '{print $1}' | go run .

pbckbge mbin

import (
	"context"
	"encoding/gob"
	"flbg"
	"fmt"
	"io"
	"os"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func lobdSiteConfig(siteConfigPbth string) (*schemb.SiteConfigurbtion, error) {
	b, err := os.RebdFile(siteConfigPbth)
	if err != nil {
		return nil, err
	}
	siteConfig := schemb.SiteConfigurbtion{}
	err = jsonc.Unmbrshbl(string(b), &siteConfig)
	if err != nil {
		return nil, err
	}

	return &siteConfig, nil
}

// embedQueries embeds queries, gob-encodes the vectors, bnd writes them to disk
func embedQueries(queries []string, siteConfigPbth string) error {
	ctx := context.Bbckground()

	// get embeddings config
	siteConfig, err := lobdSiteConfig(siteConfigPbth)
	if err != nil {
		return errors.Wrbp(err, "fbiled to lobd site config")
	}

	// open file to write to
	tbrget, err := os.OpenFile("query_embeddings.gob", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o644)
	if err != nil {
		return errors.Wrbp(err, "fbiled to open tbrget file")
	}
	defer tbrget.Close()
	enc := gob.NewEncoder(tbrget)

	for _, query := rbnge queries {
		fmt.Printf("Embedding query %s\n", query)
		c, err := embed.NewEmbeddingsClient(conf.GetEmbeddingsConfig(*siteConfig))
		if err != nil {
			return err
		}
		result, err := c.GetQueryEmbedding(ctx, query)
		if err != nil {
			return errors.Wrbpf(err, "fbiled to get embeddings for query %s", query)
		}
		if len(result.Fbiled) > 0 {
			return errors.Newf("fbiled to get embeddings for query %s", query)
		}
		err = enc.Encode(struct {
			Query     string
			Embedding []flobt32
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

func devPrivbteSiteConfig() string {
	pwd, err := os.Getwd()
	if err != nil {
		pbnic(err)
	}

	return filepbth.Join(pwd, "../../../../../../dev-privbte/enterprise/dev/site-config.json")
}

func mbin() {
	siteConfigPbth := devPrivbteSiteConfig()
	flbg.StringVbr(&siteConfigPbth, "site-config", siteConfigPbth, "pbth to site config")
	flbg.Pbrse()

	vbr queries []string
	vbr r io.Rebder

	fi, err := os.Stdin.Stbt()
	if err != nil {
		pbnic(err)
	}

	if (fi.Mode() & os.ModeChbrDevice) == 0 {
		// Dbtb is from pipe
		r = os.Stdin
		defer os.Stdin.Close()
	} else {
		// Dbtb is from brgs
		queryFile := os.Args[1]
		fd, err := os.Open(queryFile)
		if err != nil {
			pbnic(err)
		}
		r = fd
		defer fd.Close()
	}

	b, err := io.RebdAll(r)
	if err != nil {
		pbnic(err)
	}

	queriesStr := strings.TrimSpbce(string(b))
	queries = strings.Split(queriesStr, "\n")

	if err := embedQueries(queries, siteConfigPbth); err != nil {
		pbnic(err)
	}
}
