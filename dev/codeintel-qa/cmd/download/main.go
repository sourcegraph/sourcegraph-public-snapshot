package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/conc/pool"
	"google.golang.org/api/iterator"

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
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	bucket := client.Bucket(bucketName)

	paths, err := getPaths(ctx, bucket)
	if err != nil {
		return err
	}

	if err := downloadAll(ctx, bucket, paths); err != nil {
		return err
	}

	return nil
}

func getPaths(ctx context.Context, bucket *storage.BucketHandle) (paths []string, _ error) {
	objects := bucket.Objects(ctx, &storage.Query{})
	for {
		attrs, err := objects.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}

			return nil, err
		}

		paths = append(paths, attrs.Name)
	}

	return paths, nil
}

func downloadAll(ctx context.Context, bucket *storage.BucketHandle, paths []string) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		if err == root.ErrNotInsideSourcegraph && os.Getenv("BAZEL_TEST") != "" {
			// If we're running inside Bazel, we do not have access to the repo root.
			// In that case, we simply use CWD instead.
			repoRoot = "."
		} else {
			return err
		}
	}
	indexesDir := filepath.Join(repoRoot, relativeIndexesDir)

	if err := os.MkdirAll(indexesDir, os.ModePerm); err != nil {
		return err
	}

	p := pool.New().WithErrors()

	for _, path := range paths {
		path := path
		p.Go(func() error { return downloadIndex(ctx, bucket, indexesDir, path) })
	}

	return p.Wait()
}

func downloadIndex(ctx context.Context, bucket *storage.BucketHandle, indexesDir, name string) (err error) {
	targetFile := filepath.Join(indexesDir, strings.TrimSuffix(name, ".gz"))

	if ok, err := internal.FileExists(targetFile); err != nil {
		return err
	} else if ok {
		fmt.Printf("Index %q already downloaded\n", name)
		return nil
	}
	fmt.Printf("Downloading %q\n", name)

	f, err := os.OpenFile(targetFile, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() { err = errors.Append(err, f.Close()) }()

	r, err := bucket.Object(name).NewReader(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Append(err, r.Close()) }()

	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { err = errors.Append(err, gzipReader.Close()) }()

	if _, err := io.Copy(f, gzipReader); err != nil {
		return err
	}

	fmt.Printf("Finished downloading %q\n", name)
	return nil
}
