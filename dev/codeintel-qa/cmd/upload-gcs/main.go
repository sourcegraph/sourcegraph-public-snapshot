package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/dev/codeintel-qa/internal"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func main() {
	if err := mainErr(context.Background()); err != nil {
		fmt.Printf("%s error: %s\n", internal.EmojiFailure, err.Error())
		os.Exit(1)
	}
}

const (
	bucketName         = "codeintel-qa-indexes"
	relativeIndexesDir = "dev/codeintel-qa/testdata/indexes"
)

func mainErr(ctx context.Context) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	indexesDir := filepath.Join(repoRoot, relativeIndexesDir)

	names, err := getNames(ctx, indexesDir)
	if err != nil {
		return err
	}

	if err := uploadAll(ctx, indexesDir, names); err != nil {
		return err
	}

	return nil
}

func getNames(ctx context.Context, indexesDir string) (names []string, _ error) {
	entries, err := os.ReadDir(indexesDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names, nil
}

func uploadAll(ctx context.Context, indexesDir string, names []string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	bucket := client.Bucket(bucketName)

	p := pool.New().WithErrors()

	for _, name := range names {
		name := name
		p.Go(func() error { return uploadIndex(ctx, bucket, indexesDir, name) })
	}

	return p.Wait()
}

func uploadIndex(ctx context.Context, bucket *storage.BucketHandle, indexesDir, name string) (err error) {
	f, err := os.Open(filepath.Join(indexesDir, name))
	if err != nil {
		return err
	}
	defer f.Close()

	w := bucket.Object(name + ".gz").NewWriter(ctx)
	defer func() { err = errors.Append(err, w.Close()) }()

	gzipWriter := gzip.NewWriter(w)
	defer func() { err = errors.Append(err, gzipWriter.Close()) }()

	if _, err := io.Copy(gzipWriter, f); err != nil {
		return err
	}

	return nil
}
