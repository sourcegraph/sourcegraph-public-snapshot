package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const concurrency = 50
const ignoreDescendantImports = true
const ignoreSiblingDescendantImports = true
const root = "github.com/sourcegraph/sourcegraph/enterprise/cmd"

func main() {
	if err := mainErr(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func mainErr() error {
	cmd := exec.Command("go", "list", fmt.Sprintf("%s/...", root))
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	var pkgs []string
	for _, pkg := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		pkgs = append(pkgs, strings.TrimPrefix(strings.TrimPrefix(pkg, root), "/"))
	}

	imports, err := getAllImports(pkgs)
	if err != nil {
		return err
	}

	intermediatePaths := getAllIntermediatePaths(pkgs)
	sort.Strings(intermediatePaths)
	pathTree := &treeNode{
		children: map[string]*treeNode{
			"": nestPaths("", intermediatePaths),
		},
	}

	fmt.Printf("digraph deps {\n")
	fmt.Printf("    newrank=true;\n")
	writeNodes(pathTree, imports, 1)
	writeEdges(imports)
	fmt.Printf("}\n")

	return nil
}

func getAllIntermediatePaths(pkgs []string) []string {
	uniques := map[string]struct{}{}
	for _, pkg := range pkgs {
		for _, pkg := range getIntermediatePaths(pkg) {
			uniques[pkg] = struct{}{}
		}
	}

	var flattened []string
	for key := range uniques {
		flattened = append(flattened, key)
	}
	return flattened
}

func getIntermediatePaths(pkg string) []string {
	if dirname := filepath.Dir(pkg); dirname != "." {
		return append([]string{pkg}, getIntermediatePaths(dirname)...)
	}

	return []string{pkg}
}

type treeNode struct {
	children map[string]*treeNode
}

func nestPaths(prefix string, pkgs []string) *treeNode {
	nodes := map[string]*treeNode{}

outer:
	for _, pkg := range pkgs {
		// Skip self and anything not within the current prefix
		if pkg == prefix || !isParent(pkg, prefix) {
			continue
		}

		// Skip anything already claimed by this level
		for prefix := range nodes {
			if isParent(pkg, prefix) {
				continue outer
			}
		}

		nodes[pkg] = nestPaths(pkg, pkgs)
	}

	return &treeNode{nodes}
}

func isParent(child, parent string) bool {
	return parent == "" || strings.HasPrefix(child, parent+"/")
}

func getAllImports(pkgs []string) (map[string][]string, error) {
	ch := make(chan string, len(pkgs))
	for _, pkg := range pkgs {
		ch <- pkg
	}
	close(ch)

	type pair struct {
		pkg     string
		imports []string
		err     error
	}

	var wg sync.WaitGroup
	pairs := make(chan pair, len(pkgs))

	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for pkg := range ch {
				imports, err := getImports(pkg)
				pairs <- pair{pkg, imports, err}
			}
		}()
	}
	wg.Wait()
	close(pairs)

	allImports := map[string][]string{}
	for pair := range pairs {
		if err := pair.err; err != nil {
			return nil, err
		}

		allImports[pair.pkg] = pair.imports
	}

	return allImports, nil
}

func getImports(pkg string) ([]string, error) {
	cmd := exec.Command("go", "list", "-f", `{{ join .Imports "\n" }}`, fmt.Sprintf("%s/%s", root, pkg))
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var importPackages []string
	for _, importPkg := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.HasPrefix(importPkg, root) {
			importPackages = append(importPackages, strings.TrimPrefix(importPkg, root+"/"))
		}
	}

	return importPackages, nil
}

func writeNodes(node *treeNode, imports map[string][]string, level int) {
	for pkg, children := range node.children {
		if len(children.children) == 0 {
			if pkgVisible(pkg, imports) {
				fmt.Printf("%s%s [label=\"%s\"]\n", indent(level), normalize(pkg), labelize(pkg))
			}
		} else {
			fmt.Printf("%ssubgraph cluster_%s {\n", indent(level), normalize(pkg))
			fmt.Printf("%slabel=\"%s\"\n", indent(level+1), labelize(pkg))

			if pkgVisible(pkg, imports) {
				fmt.Printf("%s%s [label=\"%s\"]\n", indent(level+1), normalize(pkg), ".")
			}

			writeNodes(children, imports, level+1)
			fmt.Printf("%s}\n", indent(level))
		}
	}
}

func writeEdges(imports map[string][]string) {
	for pkg, importPkgs := range imports {
		for _, importPkg := range importPkgs {
			if !edgeVisible(pkg, importPkg) {
				continue
			}

			fmt.Printf("    %s -> %s\n", normalize(pkg), normalize(importPkg))
		}
	}
}

func pkgVisible(importPkg string, imports map[string][]string) bool {
	for _, pkg := range imports[importPkg] {
		if edgeVisible(importPkg, pkg) {
			return true
		}
	}

	for _, pkg := range importedBy(importPkg, imports) {
		if edgeVisible(pkg, importPkg) {
			return true
		}
	}
	return false
}

func edgeVisible(pkg, importPkg string) bool {
	if ignoreDescendantImports && strings.HasPrefix(importPkg, pkg) {
		return false
	}

	if ignoreSiblingDescendantImports && strings.HasPrefix(importPkg, filepath.Dir(pkg)) {
		return false
	}

	return true
}

func importedBy(pkg string, imports map[string][]string) []string {
	var deps []string
	for k, vs := range imports {
		for _, v := range vs {
			if v == pkg {
				deps = append(deps, k)
				break
			}
		}
	}

	return deps
}

func indent(level int) string {
	return strings.Repeat(" ", level*4)
}

var replacer = strings.NewReplacer("/", "_", "-", "_", ".", "_")

func labelize(pkg string) string {
	if pkg == "" {
		pkg = root
	}
	return filepath.Base(pkg)
}

func normalize(pkg string) string {
	if pkg == "" {
		pkg = root
	}
	return replacer.Replace(pkg)
}
