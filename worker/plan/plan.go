// Package plan configures the CI test plan for a repository.
package plan

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
	"golang.org/x/net/context"
	"golang.org/x/tools/godoc/vfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	httpapirouter "src.sourcegraph.com/sourcegraph/httpapi/router"
	"src.sourcegraph.com/sourcegraph/pkg/inventory"
)

// ErrNoPlan indicates that it was not possible to create a test plan
// because the repo did not have a .drone.yml file or any code files
// that this package knows how to auto-generate a plan for.
var ErrNoPlan = errors.New("missing .drone.yml file and unable to auto-generate test plan")

// CreateServer calls Create with information about the repository
// fetched from the server. It is separate from Create so that Create
// can be called locally against a local filesystem (without needing
// to hit the server).
func CreateServer(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (*droneyaml.Config, []matrix.Axis, error) {
	cl := sourcegraph.NewClientFromContext(ctx)

	// Read the existing .drone.yml file in the repo, if any.
	config, axes, found, err := readExistingConfig(ctx, repoRev)
	if err != nil {
		return nil, nil, err
	}

	// Fetch the repo inventory to see what languages are in use.
	inv, err := cl.Repos.GetInventory(ctx, &repoRev)
	if err != nil {
		return nil, nil, err
	}

	// Construct the srclib import URL.
	srclibImportURL, err := httpapirouter.URL(httpapirouter.SrclibImport, repoRev.RouteVars())
	if err != nil {
		return nil, nil, err
	}
	srclibImportURL.Path = "/.api" + srclibImportURL.Path
	if appURL := conf.AppURL(ctx); appURL == nil {
		panic("no AppURL in context (required to construct srclib import URL)")
	} else if appURL.Scheme == "" || appURL.Host == "" {
		panic("AppURL in context has no scheme or host: " + appURL.String())
	}
	srclibImportURL = conf.AppURL(ctx).ResolveReference(srclibImportURL)

	return Create(config, axes, found, inv, srclibImportURL)
}

// CreateLocal calls Create with information about the repository
// gathered from the local filesystem.
func CreateLocal(ctx context.Context, fs vfs.FileSystem) (*droneyaml.Config, []matrix.Axis, error) {
	inv, err := inventory.Scan(ctx, walkableFileSystem{fs})
	if err != nil {
		return nil, nil, err
	}

	yamlBytes, err := vfs.ReadFile(fs, ".drone.yml")
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, err
	}

	axes, err := parseMatrix(string(yamlBytes))
	if err != nil {
		return nil, nil, err
	}

	config, err := droneyaml.Parse(yamlBytes)
	if err != nil {
		return nil, nil, err
	}

	return Create(config, axes, yamlBytes != nil, inv, nil)
}

// Create creates a CI test plan. It modifies config (which may be
// provided if a partial/full config already exists).
//
// Create should execute very quickly and must not make any server API
// calls, since it can be run locally.
func Create(config *droneyaml.Config, axes []matrix.Axis, foundYAML bool, inv *inventory.Inventory, srclibImportURL *url.URL) (*droneyaml.Config, []matrix.Axis, error) {
	if !foundYAML {
		// Generate a reasonable default configuration.
		var err error
		config, axes, err = autogenerateConfig(inv)
		if err != nil {
			return nil, nil, err
		}
	}

	// Add the srclib analysis steps to the CI test plan.
	if err := configureSrclib(inv, config, axes, srclibImportURL); err != nil {
		return nil, nil, err
	}

	return config, axes, nil
}

// readExistingConfig reads the repo's existing .drone.yml
// file, if any. If no .drone.yml file is found, found is false and
// error is nil.
func readExistingConfig(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (config *droneyaml.Config, axes []matrix.Axis, found bool, err error) {
	cl := sourcegraph.NewClientFromContext(ctx)
	file, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{RepoRev: repoRev, Path: ".drone.yml"},
	})
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			err = nil
		}
		return
	}

	found = true

	config, err = droneyaml.Parse(file.Contents)
	if err != nil {
		return
	}

	axes, err = parseMatrix(string(file.Contents))
	return
}

func parseMatrix(yaml string) ([]matrix.Axis, error) {
	axes, err := matrix.Parse(yaml)
	if err != nil {
		return nil, err
	}
	if len(axes) == 0 {
		axes = append(axes, matrix.Axis{})
	}
	return axes, nil
}

func calcMatrix(m matrix.Matrix) []matrix.Axis {
	axes := matrix.Calc(m)
	if len(axes) == 0 {
		axes = append(axes, matrix.Axis{})
	}
	return axes
}

type walkableFileSystem struct{ vfs.FileSystem }

func (walkableFileSystem) Join(path ...string) string { return filepath.Join(path...) }
