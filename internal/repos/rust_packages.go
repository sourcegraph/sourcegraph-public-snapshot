package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/crates"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewRustPackagesSource returns a new RustPackagesSource from the given external service.
func NewRustPackagesSource(svc *types.ExternalService, cf *httpcli.Factory) (*DependenciesSource, error) {
	var c schema.RustPackagesConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}

	cli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	return &DependenciesSource{
		svc:        svc,
		configDeps: c.Dependencies,
		scheme:     dependencies.RustPackagesScheme,
		src:        &rustPackagesSource{client: crates.NewClient(svc.URN(), cli)},
	}, nil
}

type rustPackagesSource struct {
	client *crates.Client
}

var _ dependenciesSource = &rustPackagesSource{}

func (s *rustPackagesSource) Get(ctx context.Context, name, version string) (reposource.PackageDependency, error) {
	dep := reposource.NewRustDependency(name, version)
	// Check if crate exists or not. Crates returns a struct detailing the errors if it cannot be found.
	metaURL := fmt.Sprintf("https://crates.io/api/v1/crates/%s/%s", dep.PackageSyntax(), dep.PackageVersion())
	resp, err := s.client.Get(ctx, metaURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch crate metadata for %s with URL %s", dep.PackageManagerSyntax(), metaURL)
	}

	// Example Error Message: {"errors":[{"detail":"not found"}]}
	var errorSlice cratesError
	if err := json.Unmarshal(resp, &errorSlice); err != nil {
		return nil, errors.Wrapf(err, "error unmarshalling crates API json '%s'", string(resp))
	}

	if len(errorSlice.Errors) > 0 {
		return nil, &errorSlice
	}

	return dep, nil
}

func (rustPackagesSource) ParseDependency(dep string) (reposource.PackageDependency, error) {
	return reposource.ParseRustDependency(dep)
}

func (rustPackagesSource) ParseDependencyFromRepoName(repoName string) (reposource.PackageDependency, error) {
	return reposource.ParseRustDependencyFromRepoName(repoName)
}

type cratesError struct {
	Errors []struct {
		Detail string `json:"detail"`
	} `json:"errors"`
}

func (e *cratesError) Error() string {
	details := make([]string, 0, len(e.Errors))
	for _, detail := range e.Errors {
		details = append(details, detail.Detail)
	}

	return "crates error(s): " + strings.Join(details, ", ")
}
