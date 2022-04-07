package lockfiles

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//
// package-lock.json
//

type packageLockDependency struct {
	Version      string
	Dev          bool
	Dependencies map[string]*packageLockDependency
}

func parsePackageLockFile(r io.Reader) ([]reposource.PackageDependency, error) {
	var lockFile struct {
		Dependencies map[string]*packageLockDependency
	}

	err := json.NewDecoder(r).Decode(&lockFile)
	if err != nil {
		return nil, errors.Errorf("decode error: %w", err)
	}

	return parsePackageLockDependencies(lockFile.Dependencies)
}

func parsePackageLockDependencies(in map[string]*packageLockDependency) ([]reposource.PackageDependency, error) {
	var (
		errs errors.MultiError
		out  = make([]reposource.PackageDependency, 0, len(in))
	)

	for name, d := range in {
		dep, err := reposource.ParseNpmDependency(name + "@" + d.Version)
		if err != nil {
			errs = errors.Append(errs, err)
		} else {
			out = append(out, dep)
		}

		if d.Dependencies != nil {
			// Recursion
			deps, err := parsePackageLockDependencies(d.Dependencies)
			out = append(out, deps...)
			errs = errors.Append(errs, err)
		}
	}

	return out, errs
}

//
// yarn.lock
//

func parseYarnLockFile(r io.Reader) (deps []reposource.PackageDependency, err error) {
	var (
		name string
		skip bool
		errs errors.MultiError
	)

	/* yarn.lock

	__metadata:
	  version: 4
	  cacheKey: 6

	"asap@npm:~2.0.6":
	  version: 2.0.6
	  resolution: "asap@npm:2.0.6"
	  checksum: 3d314f8c598b625a98347bacdba609d4c889c616ca5d8ea65acaae8050ab8b7aa6630df2cfe9856c20b260b432adf2ee7a65a1021f268ef70408c70f809e3a39
	  languageName: node
	  linkType: hard
	*/

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 1 {
			continue
		}

		var version string
		if version, err = getVersion(line); err == nil { // e.g. version: 2.0.6
			if skip {
				continue
			}

			if name == "" {
				return nil, errors.New("invalid yarn.lock format")
			}

			dep, err := reposource.ParseNpmDependency(name + "@" + version)
			if err != nil {
				errs = errors.Append(errs, err)
			} else {
				deps = append(deps, dep)
				name = ""
			}
			continue
		}

		if skip = strings.HasPrefix(line, "__metadata"); skip {
			continue
		}

		if line[:1] != " " && line[:1] != "#" { // e.g. "asap@npm:~2.0.6":
			var packagename, protocol string
			if packagename, protocol, err = parsePackageLocator(line); err != nil {
				continue
			}
			if skip = !validProtocol(protocol); skip {
				continue
			}
			name = packagename
		}
	}
	return deps, errs
}

var (
	yarnLocatorRegexp = lazyregexp.New(`"?(?P<package>.+?)@(?:(?P<protocol>.+?):)?.+`)
	yarnVersionRegexp = lazyregexp.New(`\s+"?version:?"?\s+"?(?P<version>[^"]+)"?`)
)

func parsePackageLocator(target string) (packagename, protocol string, err error) {
	capture := yarnLocatorRegexp.FindStringSubmatch(target)
	if len(capture) < 2 {
		return "", "", errors.New("not package format")
	}
	for i, group := range yarnLocatorRegexp.SubexpNames() {
		switch group {
		case "package":
			packagename = capture[i]
		case "protocol":
			protocol = capture[i]
		}
	}
	return
}

func getVersion(target string) (version string, err error) {
	capture := yarnVersionRegexp.FindStringSubmatch(target)
	if len(capture) < 2 {
		return "", errors.New("not version")
	}
	return capture[len(capture)-1], nil
}

func validProtocol(protocol string) (valid bool) {
	switch protocol {
	// only scan npm packages
	case "npm", "":
		return true
	}
	return false
}
