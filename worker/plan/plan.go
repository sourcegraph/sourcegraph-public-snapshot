// Package plan configures the CI test plan for a repository.
package plan

import (
	"errors"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	httpapirouter "src.sourcegraph.com/sourcegraph/httpapi/router"
)

// ErrNoPlan indicates that it was not possible to create a test plan
// because the repo did not have a .drone.yml file or any code files
// that this package knows how to auto-generate a plan for.
var ErrNoPlan = errors.New("missing .drone.yml file and unable to auto-generate test plan")

// Create creates a CI test plan. If it is not possible to create a CI
// test plan, ErrNoPlan is returned.
func Create(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (config *droneyaml.Config, axes []matrix.Axis, err error) {
	cl := sourcegraph.NewClientFromContext(ctx)

	// Read the existing .drone.yml file in the repo, if any.
	config, axes, found, err := readExistingConfig(ctx, repoRev)
	if err != nil {
		return
	}

	inv, err := cl.Repos.GetInventory(ctx, &repoRev)
	if err != nil {
		return
	}

	if !found {
		// Generate a reasonable default configuration.
		config, axes, err = autogenerateConfig(inv)
		if err != nil {
			return
		}
	}

	// Construct the srclib import URL.
	srclibImportURL, err := httpapirouter.URL(httpapirouter.SrclibImport, repoRev.RouteVars())
	if err != nil {
		return
	}
	srclibImportURL.Path = "/.api" + srclibImportURL.Path
	if appURL := conf.AppURL(ctx); appURL == nil {
		panic("no AppURL in context (required to construct srclib import URL)")
	} else if appURL.Scheme == "" || appURL.Host == "" {
		panic("AppURL in context has no scheme or host: " + appURL.String())
	}
	srclibImportURL = conf.AppURL(ctx).ResolveReference(srclibImportURL)

	// Add the srclib analysis steps to the CI test plan.
	if err = configureSrclib(inv, config, axes, srclibImportURL); err != nil {
		return
	}

	return
}

// readExistingConfig reads and parses the repo's existing .drone.yml
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
