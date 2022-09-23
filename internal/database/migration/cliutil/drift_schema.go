package cliutil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ExpectedSchemaFactory func(filename, version string) (descriptions.SchemaDescription, bool, error)

// GCSExpectedSchemaFactory reads schema definitions from a public GCS bucket that contains schema definitions for
// a version of Sourcegraph that did not yet contain the squashed schema description file in-tree. These files
// have been backfilled to this bucket by hand. A false-valued flag is returned if the schema does not exist
// for this version.
//
// See the ./drift-schemas directory for more details on how this data was generated.
func GCSExpectedSchemaFactory(filename, version string) (schemaDescription descriptions.SchemaDescription, _ bool, _ error) {
	return fetchSchema(fmt.Sprintf("https://storage.googleapis.com/sourcegraph-assets/migrations/drift/%s-%s", version, strings.ReplaceAll(filename, "/", "_")))
}

// GitHubExpectedSchemaFactory reads schema definitions from the sourcegraph/sourcegraph repository via the
// GitHub raw API. A false-valued flag is returned if the schema does not exist for this version.
func GitHubExpectedSchemaFactory(filename, version string) (descriptions.SchemaDescription, bool, error) {
	if !regexp.MustCompile(`(^\d+.\d+$)|(^v\d+\.\d+\.\d+$)|(^[A-Fa-f0-9]{40}$)`).MatchString(version) {
		return descriptions.SchemaDescription{}, false, errors.Newf("failed to parse %q - expected a version of the form `vX.Y.Z` or a 40-character commit hash", version)
	}

	return fetchSchema(fmt.Sprintf("https://raw.githubusercontent.com/sourcegraph/sourcegraph/%s/%s", version, filename))
}

// fetchSchema makes an HTTP GET request to the given URL and reads the schema description from the response
// body. If the URL is well-formed but does not point to an existing file, a false-valued flag is returned.
func fetchSchema(url string) (schemaDescription descriptions.SchemaDescription, _ bool, _ error) {
	resp, err := http.Get(url)
	if err != nil {
		return descriptions.SchemaDescription{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return descriptions.SchemaDescription{}, false, nil
		}

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return descriptions.SchemaDescription{}, false, errors.Newf("unexpected status %d from %s: %s", resp.StatusCode, url, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(&schemaDescription); err != nil {
		return descriptions.SchemaDescription{}, false, err
	}

	return schemaDescription, true, err
}
