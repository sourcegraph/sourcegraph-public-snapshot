package server

import (
	"context"
	"go/build"
	"log"
	"time"

	"github.com/sourcegraph/go-langserver/langserver/util"
	"github.com/sourcegraph/go-langserver/pkg/tools"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

// workspacePackagesTimeout is the timeout used for workspace/xpackages
// calls.
const workspacePackagesTimeout = time.Minute

func (h *BuildHandler) handleWorkspacePackages(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request) ([]*lspext.PackageInformation, error) {
	// TODO: Add support for the cancelRequest LSP method instead of using
	// hard-coded timeouts like this here.
	//
	// See: https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#cancelRequest
	ctx, cancel := context.WithTimeout(ctx, workspacePackagesTimeout)
	defer cancel()

	rootPath := h.RootFSPath
	bctx := h.lang.BuildContext(ctx)
	findPackage := h.FindPackage
	var pkgs []*lspext.PackageInformation
	for _, pkg := range tools.ListPkgsUnderDir(bctx, rootPath) {
		bpkg, err := findPackage(ctx, bctx, pkg, rootPath, 0)
		if err != nil && !isMultiplePackageError(err) {
			log.Printf("skipping possible package %s: %s", pkg, err)
			continue
		}
		if packageIsEmpty(bpkg) {
			continue
		}
		pkgs = append(pkgs, &lspext.PackageInformation{
			Package: toPackageInformation(bpkg),
			// TODO(sqs): set pkg.Dependencies
		})
	}
	return pkgs, nil
}

func packageIsEmpty(pkg *build.Package) bool {
	return len(pkg.GoFiles) == 0 && len(pkg.CgoFiles) == 0 && len(pkg.TestGoFiles) == 0 && len(pkg.XTestGoFiles) == 0
}

func toPackageInformation(pkg *build.Package) lspext.PackageDescriptor {
	// Keep this in correspondence with goDependencyReference. The intersection of fields
	// must identify the package.
	return map[string]interface{}{
		"package": pkg.ImportPath,
		"vendor":  util.IsVendorDir(pkg.Dir),
		"doc":     pkg.Doc,
	}
}
