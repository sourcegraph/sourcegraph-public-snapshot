package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
)

// validateExtensionManifest validates a JSON extension manifest for syntax.
//
// TODO(sqs): Also validate it against the JSON Schema.
func validateExtensionManifest(text string) error {
	var o interface{}
	return jsonc.Unmarshal(text, &o)
}

// getExtensionManifestWithBundleURL returns the extension manifest as JSON. If there are no
// releases, it returns a nil manifest. If the manifest has no "url" field itself, a "url" field
// pointing to the extension's bundle is inserted. It also returns the date that the release was
// published.
func getExtensionManifestWithBundleURL(ctx context.Context, extensionID string, registryExtensionID int32, releaseTag string) (manifest *string, publishedAt time.Time, err error) {
	release, err := dbReleases{}.GetLatest(ctx, registryExtensionID, releaseTag, false)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, time.Time{}, err
	}
	if release != nil {
		// Add URL to bundle if necessary.
		var o map[string]interface{}
		if err := jsonc.Unmarshal(release.Manifest, &o); err != nil {
			return nil, time.Time{}, fmt.Errorf("parsing extension manifest for extension with ID %d (release tag %q): %s", registryExtensionID, releaseTag, err)
		}
		if o == nil {
			o = map[string]interface{}{}
		}
		urlStr, _ := o["url"].(string)
		if urlStr == "" {
			// Insert "url" field with link to bundle file on this site.
			bundleURL, err := makeExtensionBundleURL(release.ID, release.CreatedAt.UnixNano(), extensionID)
			if err != nil {
				return nil, time.Time{}, err
			}
			o["url"] = bundleURL
			b, err := json.MarshalIndent(o, "", "  ")
			if err != nil {
				return nil, time.Time{}, err
			}
			release.Manifest = string(b)
		}

		manifest = &release.Manifest
		publishedAt = release.CreatedAt
	}

	return manifest, publishedAt, nil
}

var nonLettersDigits = regexp.MustCompile(`[^a-zA-Z0-9-]`)

func makeExtensionBundleURL(registryExtensionReleaseID int64, timestamp int64, extensionIDHint string) (string, error) {
	u, err := url.Parse(conf.Get().Critical.ExternalURL)
	if err != nil {
		return "", err
	}
	extensionIDHint = nonLettersDigits.ReplaceAllString(extensionIDHint, "-") // sanitize for URL path
	u.Path = path.Join(u.Path, fmt.Sprintf("/-/static/extension/%d-%s.js", registryExtensionReleaseID, extensionIDHint))
	u.RawQuery = strconv.FormatInt(timestamp, 36) + "--" + extensionIDHint // meaningless value, just for cache-busting
	return u.String(), nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_681(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
