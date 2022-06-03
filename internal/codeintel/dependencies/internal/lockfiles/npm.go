package lockfiles

import (
	"bufio"
	"encoding/json"
	"fmt"
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

func parseYarnLockFile(r io.Reader) (deps []reposource.PackageDependency, graph *dependencyGraph, err error) {
	var (
		name string
		skip bool
		errs errors.MultiError

		current             *reposource.NpmDependency
		parsingDependencies bool
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

	// var dependencies = map[*reposource.NpmDependency][]string{}
	var byName = map[string]*reposource.NpmDependency{}
	var dependencyNames = map[*reposource.NpmDependency][]string{}

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
				return nil, nil, errors.New("invalid yarn.lock format")
			}

			dep, err := reposource.ParseNpmDependency(name + "@" + version)
			if err != nil {
				errs = errors.Append(errs, err)
			} else {
				deps = append(deps, dep)
				byName[name] = dep
				current = dep
				name = ""
			}
			continue
		}

		if skip = strings.HasPrefix(line, "__metadata"); skip {
			continue
		}

		if line[:1] != " " && line[:1] != "#" { // e.g. "asap@npm:~2.0.6":
			parsingDependencies = false

			var packagename, protocol string
			if packagename, protocol, err = parsePackageLocator(line); err != nil {
				continue
			}
			if skip = !validProtocol(protocol); skip {
				continue
			}
			name = packagename
			current = nil
		}

		if line == "  dependencies:" {
			parsingDependencies = true
		}

		if line[:4] == "    " && parsingDependencies {
			elems := strings.Split(line[4:], " ")
			name := elems[0]

			if deps, ok := dependencyNames[current]; !ok {
				dependencyNames[current] = []string{name}
			} else {
				dependencyNames[current] = append(deps, name)
			}
		}
	}

	graph = &dependencyGraph{
		roots:        make(map[*reposource.NpmDependency]struct{}),
		dependencies: make(map[*reposource.NpmDependency][]*reposource.NpmDependency),
		edges:        map[edge]struct{}{},
	}

	for pkg, depNames := range dependencyNames {
		graph.roots[pkg] = struct{}{}

		for _, depname := range depNames {
			depPkg, ok := byName[depname]
			if !ok {
				fmt.Printf("couldn't find dep by name: %q\n", depname)
				continue
			}

			if deps, ok := graph.dependencies[pkg]; !ok {
				graph.dependencies[pkg] = []*reposource.NpmDependency{depPkg}
			} else {
				graph.dependencies[pkg] = append(deps, depPkg)
			}

			graph.edges[edge{pkg, depPkg}] = struct{}{}
		}
	}

	for edge := range graph.edges {
		delete(graph.roots, edge.target)
	}

	return deps, graph, errs
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

type edge struct {
	source, target *reposource.NpmDependency
}

type dependencyGraph struct {
	roots        map[*reposource.NpmDependency]struct{}
	dependencies map[*reposource.NpmDependency][]*reposource.NpmDependency

	edges map[edge]struct{}

	nodes []*reposource.NpmDependency
}

func printGraph(graph *dependencyGraph) string {
	var out strings.Builder

	for root := range graph.roots {
		printDependencies(&out, graph, 0, root)
	}

	return out.String()
}

func printDependencies(out io.Writer, graph *dependencyGraph, level int, node *reposource.NpmDependency) {
	deps, ok := graph.dependencies[node]
	if !ok {
		fmt.Fprintf(out, "%s%s\n", strings.Repeat("\t", level), node.RepoName())
		return
	}

	fmt.Fprintf(out, "%s%s:\n", strings.Repeat("\t", level), node.RepoName())

	for _, dep := range deps {
		printDependencies(out, graph, level+1, dep)
	}
}
