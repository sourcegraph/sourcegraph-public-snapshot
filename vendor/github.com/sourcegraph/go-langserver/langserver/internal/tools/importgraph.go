// Original importgraph.Build contains the below copyright notice:
//
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tools

import (
	"go/build"
	"sync"

	"golang.org/x/tools/refactor/importgraph"
)

// FindPackageFunc is the same type as loader.Config.FindPackage. Refer to its docstring.
type FindPackageFunc func(ctxt *build.Context, fromDir, importPath string, mode build.ImportMode) (*build.Package, error)

// BuildReverseImportGraph is much like importgraph.Build, except:
// * it only returns the reverse graph
// * it does not return errors
// * it uses a custom FindPackageFunc
// * it only searches pkgs under dir (but graph can contain pkgs outside of dir)
// * it searches xtest pkgs as well
//
// The code is adapted from the original function.
func BuildReverseImportGraph(ctxt *build.Context, findPackage FindPackageFunc, dir string) importgraph.Graph {
	type importEdge struct {
		from, to string
	}

	ch := make(chan importEdge)

	go func() {
		sema := make(chan int, 20) // I/O concurrency limiting semaphore
		var wg sync.WaitGroup
		for _, path := range ListPkgsUnderDir(ctxt, dir) {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()

				sema <- 1
				// Even in error cases, Import usually returns a package.
				bp, _ := findPackage(ctxt, path, "", 0)
				<-sema

				memo := make(map[string]string)
				absolutize := func(path string) string {
					canon, ok := memo[path]
					if !ok {
						sema <- 1
						bp2, _ := findPackage(ctxt, path, bp.Dir, build.FindOnly)
						<-sema

						if bp2 != nil {
							canon = bp2.ImportPath
						} else {
							canon = path
						}
						memo[path] = canon
					}
					return canon
				}

				if bp != nil {
					for _, imp := range bp.Imports {
						ch <- importEdge{path, absolutize(imp)}
					}
					for _, imp := range bp.TestImports {
						ch <- importEdge{path, absolutize(imp)}
					}
					for _, imp := range bp.XTestImports {
						ch <- importEdge{path, absolutize(imp)}
					}
				}

			}(path)
		}
		wg.Wait()
		close(ch)
	}()

	reverse := make(importgraph.Graph)

	for e := range ch {
		if e.to == "C" {
			continue // "C" is fake
		}
		addEdge(reverse, e.to, e.from)
	}

	return reverse
}

func addEdge(g importgraph.Graph, from, to string) {
	edges := g[from]
	if edges == nil {
		edges = make(map[string]bool)
		g[from] = edges
	}
	edges[to] = true
}
