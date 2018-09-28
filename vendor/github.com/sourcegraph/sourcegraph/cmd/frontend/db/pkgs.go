package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

type PkgsProvider interface {
	UpdateIndexForLanguage(ctx context.Context, language string, repo api.RepoID, pks []lspext.PackageInformation) error
	ListPackages(ctx context.Context, op *api.ListPackagesOp) ([]*api.PackageInfo, error)
	Delete(ctx context.Context, repo api.RepoID) error
}

type pkgs struct{}

func (p *pkgs) UpdateIndexForLanguage(ctx context.Context, language string, repo api.RepoID, pks []lspext.PackageInformation) error {
	return nil
}

func (p *pkgs) ListPackages(ctx context.Context, op *api.ListPackagesOp) ([]*api.PackageInfo, error) {
	if Mocks.Pkgs.ListPackages != nil {
		return Mocks.Pkgs.ListPackages(ctx, op)
	}
	return nil, nil
}

func (p *pkgs) Delete(ctx context.Context, repo api.RepoID) error {
	if Mocks.Pkgs.Delete != nil {
		return Mocks.Pkgs.Delete(ctx, repo)
	}
	return nil
}
