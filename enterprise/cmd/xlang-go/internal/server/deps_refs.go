package server

import (
	"fmt"
	"go/build"
	"sort"
	"strings"

	"github.com/sourcegraph/go-langserver/langserver/util"
)

type goDependencyReference struct {
	pkg, absolute string
	vendor        bool
	depth         int
}

func (r *goDependencyReference) attributes() map[string]interface{} {
	// Keep this in correspondence with toPackageInformation. The intersection of fields
	// must identify the package.
	return map[string]interface{}{
		"package":  r.pkg,
		"absolute": r.absolute,
		"vendor":   r.vendor,
		"depth":    r.depth,
	}
}

// importRecord describes that pkg has imported another package.
type importRecord struct {
	pkg, imports *build.Package
}

func (i importRecord) String() string {
	return fmt.Sprintf("importRecord{pkg: %s, imports: %s}", i.pkg.ImportPath, i.imports.ImportPath)
}

type sortedImportRecord []importRecord

func (s sortedImportRecord) Len() int      { return len(s) }
func (s sortedImportRecord) Swap(i, j int) { s[j], s[i] = s[i], s[j] }
func (s sortedImportRecord) Less(i, j int) bool {
	if s[i].pkg.ImportPath != s[j].pkg.ImportPath {
		return s[i].pkg.ImportPath < s[j].pkg.ImportPath
	}
	return s[i].imports.ImportPath < s[j].imports.ImportPath
}

func unvendoredPath(abs string) string {
	i := strings.Index(abs, "/vendor/")
	if i < 0 {
		return abs
	}
	return abs[i+len("/vendor/"):]
}

// references calls emitRef on each transitive package that has been seen by
// the dependency cache. The parameters say that the Go package directory `path`
// has imported the Go package described by r.
func (d *depCache) references(emitRef func(path string, r goDependencyReference), depthLimit int) {
	// Example import graph with edge cases:
	//
	//       '/' (root)
	//        |
	//        a
	//        |\
	//        b c
	//         \|
	//    .>.   d <<<<<<.
	//    |  \ / \    | |
	//    .<< e   f >>^ |
	//        |         |
	//        f >>>>>>>>^
	//
	// Although Go does not allow such cyclic import graphs, we must handle
	// them here due to the fact that we aggregate imports for all packages in
	// a directory (e.g. including xtest files, which can import the package
	// path itself).

	// orderedEmit emits the dependency references found in m as being
	// referenced by the given path. The only difference from emitRef is that
	// the emissions are in a sorted order rather than in random map order.
	orderedEmit := func(path string, m map[string]goDependencyReference) {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			emitRef(path, m[k])
		}
	}

	// Prepare a function to walk every package node in the above example graph.
	beganWalk := map[string]struct{}{}
	var walk func(rootDir, pkgDir string, parentDirs []string, emissions map[string]goDependencyReference, depth int)
	walk = func(rootDir, pkgDir string, parentDirs []string, emissions map[string]goDependencyReference, depth int) {
		if depth >= depthLimit {
			return
		}

		// The imports are recorded in parallel by goroutines in doDeps, so we
		// must sort them in order to get a stable output order.
		imports := d.seen[pkgDir]
		sort.Sort(sortedImportRecord(imports))

		for _, imp := range imports {
			// At this point we know that `imp.pkg.ImportPath` has imported
			// `imp.imports.ImportPath`.

			// If the package being referenced is the package itself, i.e. the
			// package tried to import itself, do not walk any further.
			if imp.pkg.Dir == imp.imports.Dir {
				continue
			}

			// If the package being referenced is itself one of our parent
			// packages, then we have hit a cyclic dependency and should not
			// walk any further.
			cyclic := false
			for _, parentDir := range parentDirs {
				if parentDir == imp.imports.Dir {
					cyclic = true
					break
				}
			}
			if cyclic {
				continue
			}

			// Walk the referenced dependency so that we emit transitive
			// dependencies.
			walk(rootDir, imp.imports.Dir, append(parentDirs, pkgDir), emissions, depth+1)

			// If the dependency being referenced has not already been walked
			// individually / on its own, do so now.
			_, began := beganWalk[imp.imports.Dir]
			if !began {
				beganWalk[imp.imports.Dir] = struct{}{}
				childEmissions := map[string]goDependencyReference{}
				walk(imp.imports.Dir, imp.imports.Dir, append(parentDirs, pkgDir), childEmissions, 0)
				orderedEmit(imp.imports.Dir, childEmissions)
			}

			// If the new emissions for the import path would have a greater
			// depth, then do not overwrite the old emission. This ensures that
			// for a single package which is referenced we always get the
			// closest (smallest) depth value.
			if existing, ok := emissions[imp.imports.ImportPath]; ok {
				if existing.depth < depth {
					return
				}
			}
			emissions[imp.imports.ImportPath] = goDependencyReference{
				pkg:      unvendoredPath(imp.imports.ImportPath),
				absolute: imp.imports.ImportPath,
				vendor:   util.IsVendorDir(imp.imports.Dir),
				depth:    depth,
			}
		}
	}
	sort.Strings(d.entryPackageDirs)
	for _, entryDir := range d.entryPackageDirs {
		emissions := map[string]goDependencyReference{}
		walk(entryDir, entryDir, nil, emissions, 0)
		orderedEmit(entryDir, emissions)
	}
}
