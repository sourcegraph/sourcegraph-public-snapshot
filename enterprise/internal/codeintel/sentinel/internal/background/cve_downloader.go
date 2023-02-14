package background

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewCVEDownloader(store store.Store, metrics *Metrics, interval time.Duration) goroutine.BackgroundRoutine {
	cveDownloader := &cveDownloader{
		store: store,
	}

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.sentinel-cve-downloader", "TODO",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return cveDownloader.handle(ctx, metrics)
		}),
	)
}

type cveDownloader struct {
	store store.Store
}

const advisoryDatabaseURL = "https://github.com/github/advisory-database/archive/refs/heads/main.zip"

type Vulnerability struct {
	ID string `json:"id"`
	// TODO
}

func (matcher *cveDownloader) handle(ctx context.Context, metrics *Metrics) error {
	resp, err := http.Get(advisoryDatabaseURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Newf("unexpected status code %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		if filepath.Ext(f.Name) != ".json" {
			continue
		}

		r, err := f.Open()
		if err != nil {
			return err
		}
		defer r.Close()

		var vulnerability Vulnerability
		if err := json.NewDecoder(r).Decode(&vulnerability); err != nil {
			return err
		}

		fmt.Printf("> %v\n", vulnerability)
	}

	// TODO - insert into DB
	return nil
}
