package lockfiles

import (
	"bufio"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//
// yarn.lock
//

func parseYarnLockFile(r io.Reader) (deps []reposource.PackageVersion, graph *DependencyGraph, err error) {
	var (
		yarnLockfileV1 bool

		name        string
		constraints []string
		skip        bool
		errs        errors.MultiError

		current             *reposource.NpmPackageVersion
		parsingDependencies bool

		dependencies        = map[*reposource.NpmPackageVersion][]npmDependency{}
		dependencyToPackage = map[npmDependency]*reposource.NpmPackageVersion{}
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

		if line == "# yarn lockfile v1" {
			yarnLockfileV1 = true
		}

		var version string
		if version, err = getVersion(line); err == nil { // e.g. version: 2.0.6
			if skip {
				continue
			}

			// We've got all the data we need now to build the package.
			pkg, pkgConstraints, err := buildPackage(name, version, constraints)
			if err != nil {
				errs = errors.Append(errs, err)
				continue
			}

			deps = append(deps, pkg)
			current = pkg

			for _, constraint := range pkgConstraints {
				dependencyToPackage[constraint] = pkg
			}

			dependencies[current] = []npmDependency{}
			name = ""
			constraints = nil

			continue
		}

		if skip = strings.HasPrefix(line, "__metadata"); skip {
			continue
		}

		if line[:1] != " " && line[:1] != "#" { // e.g. "asap@npm:~2.0.6":
			parsingDependencies = false

			// Multiple protocols can be specified per line, e.g.: ajv@^6.10.0, ajv@^6.12.4:
			packageName, packageConstraints, err := parsePackageLocatorLine(line, yarnLockfileV1)
			if err != nil {
				continue
			}

			if skip = !validProtocols(packageConstraints); skip {
				continue
			}

			packageVersionConstraints := make([]string, len(packageConstraints))
			for i, pc := range packageConstraints {
				packageVersionConstraints[i] = pc.Version
			}

			name = packageName
			constraints = packageVersionConstraints
			current = nil
		}

		if line == "  dependencies:" {
			parsingDependencies = true
		}

		if line == "  bin:" {
			parsingDependencies = false
		}

		if len(line) >= 4 && line[:4] == "    " && parsingDependencies && current != nil {
			dep, err := parsePackageDependencyLine(line)
			if err != nil {
				continue
			}

			if deps, ok := dependencies[current]; !ok {
				dependencies[current] = []npmDependency{dep}
			} else {
				dependencies[current] = append(deps, dep)
			}
		}
	}

	graph = newDependencyGraph()
	for pkg, deps := range dependencies {
		graph.addPackage(pkg)

		for _, depConstraint := range deps {
			dep, ok := dependencyToPackage[depConstraint]
			if !ok {
				errs = errors.Append(errs, errors.Newf("could not find dependency with name %s and constraint %s", depConstraint.RepoName(), depConstraint.VersionConstraint))
			}

			graph.addDependency(pkg, dep)
		}
	}

	return deps, graph, errs
}

var (
	yarnLocatorRegexp    = lazyregexp.New(`"?(?P<package>.+?)@((?P<protocol>\w+):)?(?P<constraint>[^"]+)"?`)
	yarnDependencyRegexp = lazyregexp.New(`\s{4}"?(?P<package>.+?)"?\s"?(?P<version>[^"]+)"?`)
	yarnVersionRegexp    = lazyregexp.New(`\s+"?version:?"?\s+"?(?P<version>[^"]+)"?`)
)

type constraint struct {
	Protocol string // e.g. "npm" in lockfile v2 files
	Version  string // e.g. "~1.2.3 || 1.2.0"
}

func buildPackage(name, version string, constraints []string) (*reposource.NpmPackageVersion, []npmDependency, error) {
	if name == "" {
		return nil, nil, errors.New("invalid yarn.lock format: version not following a name")
	}

	if constraints == nil {
		return nil, nil, errors.New("invalid yarn.lock format: version not following a name with constraints")
	}

	pkg, err := reposource.ParseNpmPackageVersion(name + "@" + version)
	if err != nil {
		return nil, nil, err
	}

	var deps []npmDependency
	var errs errors.MultiError
	for _, proto := range constraints {
		dep, err := parseNpmDependency(name, proto)
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		}
		deps = append(deps, dep)
	}

	return pkg, deps, errs
}

func parsePackageLocatorLine(line string, yarnLockfileV1 bool) (packagename string, constraints []constraint, err error) {
	var elems []string
	// yarn lockfile v1 locator line looks like this (without intendation):
	//
	//    "@types/istanbul-lib-coverage@*", "@types/istanbul-lib-coverage@^2.0.0":
	//
	// But the quotes are optional, so we leave them here and handle them in
	// the regex.
	if yarnLockfileV1 {
		trimmed := strings.TrimSuffix(line, ":")
		elems = strings.Split(trimmed, ", ")
	} else {
		// yarn lockfile v2+ locator line looks like this (without intendation):
		//
		//     "console-control-strings@npm:^1.0.0, console-control-strings@npm:~1.1.0":
		//
		// We strip the quotes here, so each element can be handled by the
		// v1-compatible regex further down.
		trimmed := strings.Trim(strings.TrimSuffix(line, ":"), `"`)
		elems = strings.Split(trimmed, ", ")
	}

	for _, elem := range elems {
		capture := yarnLocatorRegexp.FindStringSubmatch(elem)
		if len(capture) < 2 {
			return "", constraints, errors.Newf("not package locator format: %s", elem)
		}
		var protocol string
		for i, group := range yarnLocatorRegexp.SubexpNames() {
			switch group {
			case "package":
				packagename = capture[i]
			case "protocol":
				protocol = capture[i]
			case "constraint":
				constraints = append(constraints, constraint{Protocol: protocol, Version: capture[i]})
			}
		}
	}

	return
}

func parsePackageDependencyLine(line string) (dep npmDependency, err error) {
	capture := yarnDependencyRegexp.FindStringSubmatch(line)
	if len(capture) < 2 {
		return npmDependency{}, errors.Newf("not package dependency format: %s", line)
	}

	var dependencyname, version string
	for i, group := range yarnDependencyRegexp.SubexpNames() {
		switch group {
		case "package":
			dependencyname = capture[i]
		case "version":
			version = capture[i]
		}
	}

	return parseNpmDependency(dependencyname, version)
}

// npmDependency is a package that another package depends on. It
// includes a VersionConstraint that specifies which version the dependent
// requires.
type npmDependency struct {
	reposource.NpmPackageName

	// The version constraint (such as "^4.6.0" or "> 12.0 || < 15") for a dependency.
	VersionConstraint string
}

func (n *npmDependency) Equal(o *npmDependency) bool {
	return n == o || (n != nil && o != nil &&
		n.NpmPackageName == o.NpmPackageName &&
		n.VersionConstraint == o.VersionConstraint)
}

// parseNpmDependency parses a dependency name (with optional scope) and
// a version constraint into a NpmPackageDependency.
func parseNpmDependency(dependency, constraint string) (npmDependency, error) {
	name, err := reposource.ParseNpmPackageNameWithoutVersion(dependency)
	if err != nil {
		return npmDependency{}, err
	}
	return npmDependency{NpmPackageName: name, VersionConstraint: constraint}, nil
}

func getVersion(target string) (version string, err error) {
	capture := yarnVersionRegexp.FindStringSubmatch(target)
	if len(capture) < 2 {
		return "", errors.Newf("not in version format: %s", target)
	}
	return capture[len(capture)-1], nil
}

func validProtocols(constraints []constraint) bool {
	for _, c := range constraints {
		if !validProtocol(c.Protocol) {
			return false
		}
	}
	return true
}

func validProtocol(protocol string) (valid bool) {
	switch protocol {
	// only scan npm packages
	case "npm", "":
		return true
	}
	return false
}
