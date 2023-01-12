// Program freeze_legacy_extensions downloads a copy of extension registry data from sourcegraph.com
// and freezes it in place so that it can be statically served from the same endpoint without
// requiring a live database anymore. It is part of our removal of the legacy Sourcegraph extension
// API.
//
// This program expects Google Cloud application default credentials set up. Run:
//
//   gcloud auth application-default login
//
// See https://cloud.google.com/docs/authentication/application-default-credentials#personal for
// more information.
//
// Run it as:
//
//   go run ./enterprise/cmd/frontend/internal/registry/scripts/freeze_legacy_extensions.go > enterprise/cmd/frontend/internal/registry/frozen_legacy_extensions.json
//
// Note: In case it's helpful, run this to get the list of extensions from sourcegraph.com's API:
//
//   curl -v -H 'Accept: application/vnd.sourcegraph.api+json;version=20180621' https://sourcegraph.com/.api/registry/extensions

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"google.golang.org/api/googleapi"
)

var (
	gcsBucket  = flag.String("bucket", "sourcegraph-legacy-extensions", "Google Cloud Storage bucket where data is uploaded")
	gcpProject = flag.String("project", "sourcegraph-legacy-extensions", "Google Cloud Platform project in which to create the bucket")
	devFirst   = flag.Int("dev-first", 0, "DEV ONLY: only operate on the first N extensions")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	log.Println("# Listing extensions from sourcegraph.com...")
	extensions, err := listExtensions()
	if err != nil {
		log.Fatal(err)
	}

	total := len(extensions)
	extensions = filterExtensions(extensions)
	log.Printf("# Freezing %d extensions (out of %d total)", len(extensions), total)

	if *devFirst > 0 {
		log.Printf("# DEV: Only operating on the first %d extensions", *devFirst)
		extensions = extensions[:*devFirst]
	}

	log.Printf("# Fetching extension bundles...")
	jsBundles := make([][]byte, len(extensions))
	for i, x := range extensions {
		var err error
		jsBundles[i], err = getExtensionJSBundle(x)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("# Uploading extension JS bundles to Google Cloud Storage...")
	jsBundleURLs, err := uploadJSBundles(extensions, jsBundles, *gcpProject, *gcsBucket)
	if err != nil {
		log.Fatal(err)
	}

	// Update extension manifests with the new URL.
	for i := range extensions {
		extensions[i].Manifest.URL = jsBundleURLs[i]
	}

	out, err := json.MarshalIndent(extensions, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}

func uploadJSBundles(extensions []extension, jsBundles [][]byte, projectName, bucketName string) (jsBundleURLs []string, err error) {
	ctx := context.Background()
	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to initialize Google Cloud Storage client")
	}

	bkt := gcsClient.Bucket(bucketName)
	if err := bkt.Create(ctx, projectName, &storage.BucketAttrs{PredefinedACL: "private"}); err != nil {
		if e, ok := err.(*googleapi.Error); ok && e.Code == http.StatusConflict {
			// Bucket already exists; ignore.
		} else {
			return nil, errors.WithMessage(err, "failed to create GCS bucket")
		}
	}

	jsBundleURLs = make([]string, len(extensions))
	for i, x := range extensions {
		jsBundle := jsBundles[i]

		log.Printf("# - %s (%.1f kb)", x.ExtensionID, float64(len(jsBundle))/1024)
		var err error
		jsBundleURLs[i], err = uploadJSBundle(ctx, bkt, x, jsBundle)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to upload JS bundle")
		}
	}

	return jsBundleURLs, nil
}

func uploadJSBundle(ctx context.Context, bkt *storage.BucketHandle, x extension, jsBundle []byte) (string, error) {
	// Change prefix to something else if you want to rerun this script and not get the old and new
	// GCS objects confused.
	const prefix = "202212"

	name := prefix + "-" + strings.Replace(x.ExtensionID, "/", "-", 1) + ".js"
	obj := bkt.Object(name)
	w := obj.NewWriter(ctx)
	if _, err := w.Write(jsBundle); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	if _, err := obj.Update(ctx, storage.ObjectAttrsToUpdate{
		PredefinedACL: "publicRead",
		ContentType:   "application/javascript",
		CacheControl:  "max-age=86400",
	}); err != nil {
		return "", err
	}

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return "", err
	}

	url := "https://storage.googleapis.com/" + attrs.Bucket + "/" + attrs.Name
	return url, nil
}

type extension struct {
	Name        string             `json:"name,omitempty"`
	ExtensionID string             `json:"extensionID,omitempty"`
	UUID        string             `json:"uuid,omitempty"`
	Publisher   publisher          `json:"publisher"`
	CreatedAt   time.Time          `json:"createdAt,omitempty"`
	UpdatedAt   time.Time          `json:"updatedAt,omitempty"`
	PublishedAt time.Time          `json:"publishedAt,omitempty"`
	Manifest    *extensionManifest `json:"manifest,omitempty"`
}

type publisher struct {
	Name string `json:"name,omitempty"`
}

type extensionManifest struct {
	ActivationEvents []string        `json:"activationEvents,omitempty"`
	Publisher        string          `json:"publisher"`
	Contributes      json.RawMessage `json:"contributes,omitempty"`
	URL              string          `json:"url"`
	Categories       []string        `json:"categories,omitempty"`
}

func (m *extensionManifest) MarshalJSON() ([]byte, error) {
	type wrapper extensionManifest
	b, err := json.Marshal((*wrapper)(m))
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(b))
}

func (m *extensionManifest) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	type wrapper extensionManifest
	var tmp wrapper
	if err := jsonc.Unmarshal(s, &tmp); err != nil {
		return err
	}
	*m = extensionManifest(tmp)
	return nil
}

func listExtensions() (extensions []extension, err error) {
	err = httpGet("/extensions", &extensions)
	return extensions, err
}

func filterExtensions(extensions []extension) []extension {
	keepExtension := func(x extension) bool {
		// Includes notable extensions from https://sourcegraph.looker.com/dashboards/371.
		return x.Publisher.Name == "sourcegraph" || x.ExtensionID == "dymka/open-in-webstorm" || x.ExtensionID == "juliosueiras/terraform-extension"
	}

	keep := extensions[:0]
	for _, x := range extensions {
		if keepExtension(x) {
			keep = append(keep, x)
		}
	}
	return keep
}

func httpGet(urlPath string, result any) error {
	req, err := http.NewRequest("GET", "https://sourcegraph.com/.api/registry"+urlPath, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.sourcegraph.api+json;version=20180621")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("http status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(&result)
}

func getExtensionJSBundle(x extension) ([]byte, error) {
	resp, err := http.Get(x.Manifest.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
