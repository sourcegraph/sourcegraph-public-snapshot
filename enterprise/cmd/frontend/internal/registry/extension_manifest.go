package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strconv"

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
// pointing to the extension's bundle is inserted.
func getExtensionManifestWithBundleURL(ctx context.Context, extensionID string, registryExtensionID int32, releaseTag string) (*string, error) {
	var manifest *string
	release, err := dbReleases{}.GetLatest(ctx, registryExtensionID, releaseTag, false)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}
	if release != nil {
		// Add URL to bundle if necessary.
		var o map[string]interface{}
		if err := jsonc.Unmarshal(release.Manifest, &o); err != nil {
			return nil, fmt.Errorf("parsing extension manifest for extension with ID %d (release tag %q): %s", registryExtensionID, releaseTag, err)
		}
		if o == nil {
			o = map[string]interface{}{}
		}
		urlStr, _ := o["url"].(string)
		if urlStr == "" {
			// Insert "url" field with link to bundle file on this site.
			bundleURL, err := makeExtensionBundleURL(release.ID, release.CreatedAt.UnixNano(), extensionID)
			if err != nil {
				return nil, err
			}
			o["url"] = bundleURL
			b, err := json.MarshalIndent(o, "", "  ")
			if err != nil {
				return nil, err
			}
			release.Manifest = string(b)
		}

		manifest = &release.Manifest
	}

	return manifest, nil
}

func makeExtensionBundleURL(registryExtensionReleaseID int64, timestamp int64, extensionIDHint string) (string, error) {
	u, err := url.Parse(conf.Get().AppURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("/-/static/extension/%d.js", registryExtensionReleaseID))
	u.RawQuery = extensionIDHint + "--" + strconv.FormatInt(timestamp, 36) // meaningless value, just for cache-busting
	return u.String(), nil
}
