package cliutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ExpectedSchemaFactory converts the given filename and version into a schema description, calling on some
// external persistent source. When invoked, this function should return a self-describing name, which notably
// should include _where_ the factory looked for easier debugging on failure, the schema description, and any
// error that occurred. Name should be returned on error when possible as well for logging purposes.
type ExpectedSchemaFactory func(filename, version string) (string, descriptions.SchemaDescription, error)

// GitHubExpectedSchemaFactory reads schema definitions from the sourcegraph/sourcegraph repository via the
// GitHub raw API.
func GitHubExpectedSchemaFactory(filename, version string) (string, descriptions.SchemaDescription, error) {
	path, err := makeRepoRevPath(filename, version)
	if err != nil {
		return "github://{UNSUPPORTED_VERSION}", descriptions.SchemaDescription{}, err
	}

	schemaDescription, err := fetchSchema(fmt.Sprintf("https://raw.githubusercontent.com/%s", path))
	return fmt.Sprintf("github://%s", path), schemaDescription, err
}

func makeRepoRevPath(filename, version string) (string, error) {
	if !versionBranchOrVersionTagOrCommitPattern.MatchString(version) {
		return "", errors.Newf("failed to parse %q - expected a version matching a Git revlike", version)
	}

	return fmt.Sprintf("sourcegraph/sourcegraph/%s/%s", version, filename), nil
}

// GCSExpectedSchemaFactory reads schema definitions from a public GCS bucket that contains schema definitions for
// a version of Sourcegraph that did not yet contain the squashed schema description file in-tree. These files
// have been backfilled to this bucket by hand.
//
// See the ./drift-schemas directory for more details on how this data was generated.
func GCSExpectedSchemaFactory(filename, version string) (string, descriptions.SchemaDescription, error) {
	path, err := makeFilePath(filename, version)
	if err != nil {
		return "gcs://{UNSUPPORTED_VERSION}", descriptions.SchemaDescription{}, err
	}

	schemaDescription, err := fetchSchema(fmt.Sprintf("https://storage.googleapis.com/sourcegraph-assets/migrations/drift/%s", path))
	return fmt.Sprintf("gcs://%s", path), schemaDescription, err
}

const migratorImageDescriptionPrefix = "/schema-descriptions"

// LocalExpectedSchemaFactory reads schema definitions from a local directory baked into the migrator image.
func LocalExpectedSchemaFactory(filename, version string) (string, descriptions.SchemaDescription, error) {
	path, err := makeFilePath(filename, version)
	if err != nil {
		return "file://{UNSUPPORTED_VERSION}", descriptions.SchemaDescription{}, err
	}

	schemaDescription, err := readSchemaFromFile(filepath.Join(migratorImageDescriptionPrefix, path))
	return fmt.Sprintf("file://%s/%s", migratorImageDescriptionPrefix, path), schemaDescription, err
}

var (
	versionBranchPattern                     = lazyregexp.New(`\d+\.\d+`)        // Versioned branches in Git have the form `{MAJOR}.{MINOR}`
	tagPattern                               = lazyregexp.New(`v\d+\.\d+\.\d+`)  // Versioned tags in Git have the form `v{MAJOR}.{MINOR}.{PATCH}`
	commitPattern                            = lazyregexp.New(`[0-9A-Fa-f]{40}`) // Commits must be the full 40 character SHA
	onlyTagPattern                           = lazyregexp.New(`^` + tagPattern.Re().String() + `$`)
	versionBranchOrVersionTagOrCommitPattern = lazyregexp.New(`` +
		`^(` + versionBranchPattern.Re().String() + `)` +
		`|(` + tagPattern.Re().String() + `)` +
		`|(` + commitPattern.Re().String() + `)` +
		`$`,
	)
)

func makeFilePath(filename, version string) (string, error) {
	if !onlyTagPattern.MatchString(version) {
		return "", errors.Newf("failed to parse %q - expected a version of the form `vX.Y.Z`", version)
	}

	return fmt.Sprintf("%s-%s", version, strings.ReplaceAll(filename, "/", "_")), nil
}

// ExplicitFileSchemaFactory creates a schema factory that reads a schema description from the given filename.
// The parameters of the returned function are ignored on invocation.
func ExplicitFileSchemaFactory(filename string) func(filename, version string) (string, descriptions.SchemaDescription, error) {
	return func(_, _ string) (string, descriptions.SchemaDescription, error) {
		schemaDescription, err := readSchemaFromFile(filename)
		return fmt.Sprintf("file://%s", filename), schemaDescription, err
	}
}

// fetchSchema makes an HTTP GET request to the given URL and reads the schema description from the response.
func fetchSchema(url string) (descriptions.SchemaDescription, error) {
	resp, err := http.Get(url)
	if err != nil {
		return descriptions.SchemaDescription{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return descriptions.SchemaDescription{}, errors.Newf("HTTP %d: %s", resp.StatusCode, url)
	}

	var schemaDescription descriptions.SchemaDescription
	err = json.NewDecoder(resp.Body).Decode(&schemaDescription)
	return schemaDescription, err
}

// readSchemaFromFile reads a schema description from the given filename.
func readSchemaFromFile(filename string) (descriptions.SchemaDescription, error) {
	f, err := os.Open(filename)
	if err != nil {
		return descriptions.SchemaDescription{}, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = errors.Append(err, closeErr)
		}
	}()

	var schemaDescription descriptions.SchemaDescription
	err = json.NewDecoder(f).Decode(&schemaDescription)
	return schemaDescription, err
}
