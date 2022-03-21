package lockfiles

import (
	"bufio"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//
// go.sum
//

func parseGoSumFile(r io.Reader) ([]reposource.PackageDependency, error) {
	var (
		errs errors.MultiError

		// In some cases, two checksums occur for both the package itself and the go.mod
		// file of the same version.
		deps = make(map[string]reposource.PackageDependency)
	)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		s := strings.Fields(line)
		if len(s) < 3 {
			continue
		}

		dep, err := reposource.ParseGoModDependency(line)
		if err != nil {
			errs = errors.Append(errs, err)
		} else {
			deps[dep.PackageManagerSyntax()] = dep
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scan")
	}

	out := make([]reposource.PackageDependency, 0, len(deps))
	for _, dep := range deps {
		out = append(out, dep)
	}
	return out, nil
}
