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
		// file of the same version, as well as merging multiple go.sum files, thus we
		// need to do deduplication.
		deps  []reposource.PackageDependency
		added = make(map[string]struct{})
	)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) < 2 { // We do not consume the checksum so not required to be presented for now
			continue
		}
		name := fields[0]
		version := strings.TrimSuffix(fields[1], "/go.mod")

		dep, err := reposource.ParseGoModDependency(name + "@" + version)
		if err != nil {
			errs = errors.Append(errs, err)
		} else if _, ok := added[dep.PackageManagerSyntax()]; !ok {
			deps = append(deps, dep)
			added[dep.PackageManagerSyntax()] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scan")
	}
	return deps, nil
}
