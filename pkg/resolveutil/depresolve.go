package resolveutil

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/golang/gddo/gosrc"
	"golang.org/x/tools/go/vcs"
)

// ResolvedTarget represents a resolved dependency target.
type ResolvedTarget struct {
	// ToRepoCloneURL is the clone URL of the repository that is depended on.
	//
	// When graphers emit ResolvedDependencies, they should fill in this field,
	// not ToRepo, so that the dependent repository can be added if it doesn't
	// exist. The ToRepo URI alone does not specify enough information to add
	// the repository (because it doesn't specify the VCS type, scheme, etc.).
	ToRepoCloneURL string

	// ToUnit is the name of the source unit that is depended on.
	ToUnit string

	// ToUnitType is the type of the source unit that is depended on.
	ToUnitType string

	// ToVersion is the version of the dependent repository (if known),
	// according to whatever version string specifier is used by FromRepo's
	// dependency management system.
	ToVersionString string

	// ToRevSpec specifies the desired VCS revision of the dependent repository
	// (if known).
	ToRevSpec string
}

func ResolveImportPath(importPath string) (*ResolvedTarget, error) {
	// Handle some special (and edge) cases faster for performance and corner-cases.
	target := &ResolvedTarget{ToUnit: importPath, ToUnitType: "GoPackage"}
	switch {
	// CGO package "C"
	case importPath == "C":
		return nil, nil

	// Go standard library packages
	case gosrc.IsGoRepoPath(importPath) || strings.HasPrefix(importPath, "debug/") || strings.HasPrefix(importPath, "cmd/"):
		target.ToRepoCloneURL = "https://github.com/golang/go"
		target.ToVersionString = runtime.Version()
		target.ToRevSpec = "" // TODO(sqs): fill in when graphing stdlib repo

	// Special-case github.com/... import paths for performance.
	case strings.HasPrefix(importPath, "github.com/"):
		cloneURL, err := standardRepoHostImportPathToCloneURL(importPath)
		if err != nil {
			return nil, err
		}
		target.ToRepoCloneURL = cloneURL

	// Special-case google.golang.org/grpc/... import paths for performance
	// and to avoid hitting GitHub rate limit. It does not follow the same pattern
	// as the general google.golang.org/... case below.
	case importPath == "google.golang.org/grpc" || strings.HasPrefix(importPath, "google.golang.org/grpc/"):
		target.ToRepoCloneURL = "https://github.com/grpc/grpc-go"
		target.ToUnit = strings.Replace(importPath, "google.golang.org/grpc", "github.com/grpc/grpc-go", 1)

	// Special-case google.golang.org/cloud/...
	// The canonical import path is hosted on https://code.googlesource.com/gocloud,
	// but sourcegraph.com uses its github.com mirror
	case importPath == "google.golang.org/cloud" || strings.HasPrefix(importPath, "google.golang.org/cloud/"):
		target.ToRepoCloneURL = "https://github.com/GoogleCloudPlatform/gcloud-golang"
		target.ToUnit = strings.Replace(importPath, "google.golang.org/cloud", "github.com/GoogleCloudPlatform/gcloud-golang", 1)

	// Special-case google.golang.org/api/...
	// The canonical import path is hosted on https://code.googlesource.com/google-api-go-client
	// but sourcegraph.com uses its github.com mirror
	case importPath == "google.golang.org/api" || strings.HasPrefix(importPath, "google.golang.org/api/"):
		target.ToRepoCloneURL = "https://github.com/google/google-api-go-client"
		target.ToUnit = strings.Replace(importPath, "google.golang.org/api", "github.com/google/google-api-go-client", 1)

	// Special-case code.google.com/p/... import paths for performance.
	case strings.HasPrefix(importPath, "code.google.com/p/"):
		parts := strings.SplitN(importPath, "/", 4)
		if len(parts) < 3 {
			return nil, fmt.Errorf("import path starts with 'code.google.com/p/' but is not valid: %q", importPath)
		}
		target.ToRepoCloneURL = "https://" + strings.Join(parts[:3], "/")

	// Special-case golang.org/x/... import paths for performance.
	case strings.HasPrefix(importPath, "golang.org/x/"):
		parts := strings.SplitN(importPath, "/", 4)
		if len(parts) < 3 {
			return nil, fmt.Errorf("import path starts with 'golang.org/x/' but is not valid: %q", importPath)
		}
		target.ToRepoCloneURL = "https://" + strings.Replace(strings.Join(parts[:3], "/"), "golang.org/x/", "github.com/golang/", 1)
		target.ToUnit = strings.Replace(importPath, "golang.org/x/", "github.com/golang/", 1)

	// Try to resolve everything else
	default:
		repoRoot, err := vcs.RepoRootForImportPath(string(importPath), false)
		if err == nil {
			target.ToRepoCloneURL = repoRoot.Repo
			target.ToUnit = replaceImportPathRepoRoot(target.ToUnit, repoRoot.Root, repoRoot.Repo)
		} else {
			log.Printf("warning: unable to fetch information about Go package %q: %s", importPath, err)
			target.ToRepoCloneURL = importPath
		}
	}
	return target, nil
}

// standardRepoHostImportPathToCloneURL returns the clone URL for an
// import path that references a standard repo host (e.g.,
// github.com). It assumes a structure of
// $HOST/$OWNER/$REPO/$PACKAGE_PATH. E.g., "github.com/foo/bar/path/to/pkg".
func standardRepoHostImportPathToCloneURL(importPath string) (string, error) {
	parts := strings.SplitN(importPath, "/", 4)
	if len(parts) < 3 {
		return "", fmt.Errorf("import path expected to have at least 3 parts, but didn't: %q", importPath)
	}
	return "https://" + strings.Join(parts[:3], "/"), nil
}

// replaceImportPathRepoRoot modifies the given importPath by replacing the
// root string in the importPath with the host+path of the clone URL.
// This is necessary for resolving refs to custom import path packages, since
// the defs within those packages would have the Unit field set to
// "${repoURI}/path/to/pkg/dir", where repoURI is the host+path of the cloneURL.
//
// This is a HACK to make def resolution work in presence of custom import
// paths. A proper solution would require passing in the custom import path
// information from srclib to the graph step, and setting that as the Unit field
// of the defs identified in the source unit.
//
// TODO: Implement the proper solution and get rid of this hack.
func replaceImportPathRepoRoot(importPath, root, cloneURL string) string {
	i := strings.Index(cloneURL, "://")
	if i < 0 {
		return importPath
	}
	newRoot := strings.TrimSuffix(cloneURL[i+len("://"):], ".git")
	return strings.Replace(importPath, root, newRoot, 1)
}
