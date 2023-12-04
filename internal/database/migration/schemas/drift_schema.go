package schemas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DefaultSchemaFactories is a list of schema factories to be used in
// non-exceptional cases.
var DefaultSchemaFactories = []ExpectedSchemaFactory{
	LocalExpectedSchemaFactory,
	GitHubExpectedSchemaFactory,
	GCSExpectedSchemaFactory,
}

// ExpectedSchemaFactory converts the given filename and version into a schema description, calling on some
// external persistent source. When invoked, this function should return a self-describing name, which notably
// should include _where_ the factory looked for easier debugging on failure, the schema description, and any
// error that occurred. Name should be returned on error when possible as well for logging purposes.
type ExpectedSchemaFactory interface {
	Name() string
	VersionPatterns() []NamedRegexp
	ResourcePath(filename, version string) string
	CreateFromPath(ctx context.Context, path string) (SchemaDescription, error)
}

type NamedRegexp struct {
	*lazyregexp.Regexp
	example string
}

func (r NamedRegexp) Example() string {
	return r.example
}

var (
	versionBranchPattern     = NamedRegexp{lazyregexp.New(`^\d+\.\d+$`), `4.1 (version branch)`}
	tagPattern               = NamedRegexp{lazyregexp.New(`^v\d+\.\d+\.\d+$`), `v4.1.1 (tagged release)`}
	commitPattern            = NamedRegexp{lazyregexp.New(`^[0-9A-Fa-f]{40}$`), `57b1f56787619464dc62f469127d64721b428b76 (40-character sha)`}
	abbreviatedCommitPattern = NamedRegexp{lazyregexp.New(`^[0-9A-Fa-f]{12}$`), `57b1f5678761 (12-character sha)`}
	allPatterns              = []NamedRegexp{versionBranchPattern, tagPattern, commitPattern, abbreviatedCommitPattern}
)

// GitHubExpectedSchemaFactory reads schema definitions from the sourcegraph/sourcegraph repository via GitHub's API.
var GitHubExpectedSchemaFactory = NewExpectedSchemaFactory("GitHub", allPatterns, GithubExpectedSchemaPath, fetchSchema)

func GithubExpectedSchemaPath(filename, version string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/sourcegraph/sourcegraph/%s/%s", version, filename)
}

// GCSExpectedSchemaFactory reads schema definitions from a public GCS bucket that contains schema definitions for
// a version of Sourcegraph that did not yet contain the squashed schema description file in-tree. These files have
// been backfilled to this bucket by hand.
//
// See the ./drift-schemas directory for more details on how this data was generated.
var GCSExpectedSchemaFactory = NewExpectedSchemaFactory("GCS", []NamedRegexp{tagPattern}, GcsExpectedSchemaPath, fetchSchema)

func GcsExpectedSchemaPath(filename, version string) string {
	return fmt.Sprintf("https://storage.googleapis.com/sourcegraph-assets/migrations/drift/%s-%s", version, strings.ReplaceAll(filename, "/", "_"))
}

// LocalExpectedSchemaFactory reads schema definitions from a local directory baked into the migrator image.
var LocalExpectedSchemaFactory = NewExpectedSchemaFactory("Local file", []NamedRegexp{tagPattern}, LocalSchemaPath, ReadSchemaFromFile)

const migratorImageDescriptionPrefix = "/schema-descriptions"

func LocalSchemaPath(filename, version string) string {
	return filepath.Join(migratorImageDescriptionPrefix, fmt.Sprintf("%s-%s", version, strings.ReplaceAll(filename, "/", "_")))
}

// NewExplicitFileSchemaFactory creates a schema factory that reads a schema description from the given filename.
// The parameters of the returned function are ignored on invocation.
func NewExplicitFileSchemaFactory(filename string) ExpectedSchemaFactory {
	return NewExpectedSchemaFactory("Local file", nil, func(_, _ string) string { return filename }, ReadSchemaFromFile)
}

// fetchSchema makes an HTTP GET request to the given URL and reads the schema description from the response.
func fetchSchema(ctx context.Context, url string) (SchemaDescription, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return SchemaDescription{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SchemaDescription{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SchemaDescription{}, errors.Newf("HTTP %d: %s", resp.StatusCode, url)
	}

	var schemaDescription SchemaDescription
	err = json.NewDecoder(resp.Body).Decode(&schemaDescription)
	return schemaDescription, err
}

// ReadSchemaFromFile reads a schema description from the given filename.
func ReadSchemaFromFile(ctx context.Context, filename string) (SchemaDescription, error) {
	f, err := os.Open(filename)
	if err != nil {
		return SchemaDescription{}, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = errors.Append(err, closeErr)
		}
	}()

	var schemaDescription SchemaDescription
	err = json.NewDecoder(f).Decode(&schemaDescription)
	return schemaDescription, err
}

//
//

type expectedSchemaFactory struct {
	name               string
	versionPatterns    []NamedRegexp
	resourcePathFunc   func(filename, version string) string
	createFromPathFunc func(ctx context.Context, path string) (SchemaDescription, error)
}

func NewExpectedSchemaFactory(
	name string,
	versionPatterns []NamedRegexp,
	resourcePathFunc func(filename, version string) string,
	createFromPathFunc func(ctx context.Context, path string) (SchemaDescription, error),
) ExpectedSchemaFactory {
	return &expectedSchemaFactory{
		name:               name,
		versionPatterns:    versionPatterns,
		resourcePathFunc:   resourcePathFunc,
		createFromPathFunc: createFromPathFunc,
	}
}

func (f expectedSchemaFactory) Name() string {
	return f.name
}

func (f expectedSchemaFactory) VersionPatterns() []NamedRegexp {
	return f.versionPatterns
}

func (f expectedSchemaFactory) ResourcePath(filename, version string) string {
	return f.resourcePathFunc(filename, version)
}

func (f expectedSchemaFactory) CreateFromPath(ctx context.Context, path string) (SchemaDescription, error) {
	return f.createFromPathFunc(ctx, path)
}
