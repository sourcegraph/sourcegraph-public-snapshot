package types

import (
	"fmt"
	"go/ast"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/derision-test/go-mockgen/v2/internal"	
	"github.com/derision-test/go-mockgen/v2/internal/mockgen/paths"
	"golang.org/x/tools/go/packages"
)

type PackageOptions struct {
	ImportPaths []string
	Interfaces  []string
	Exclude     []string
	Prefix      string

	StdlibRoot  string
	SourceFiles []string
	Archives    []string
}

func Extract(pkgs []*internal.GoPackage, packageOptions []PackageOptions) (ifaces []*Interface, _ error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory (%s)", err.Error())
	}

	for _, packageOpts := range packageOptions {
		packageTypes, err := gatherAllPackageTypes(pkgs, workingDirectory, packageOpts.ImportPaths)
		if err != nil {
			return nil, err
		}

		for _, name := range gatherAllPackageTypeNames(packageTypes) {
			iface, err := extractInterface(packageTypes, name, packageOpts.Interfaces, packageOpts.Exclude)
			if err != nil {
				return nil, err
			}

			if iface != nil {
				iface.Prefix = packageOpts.Prefix
				ifaces = append(ifaces, iface)
			}
		}
	}

	return ifaces, nil
}

func gatherAllPackageTypes(pkgs []*internal.GoPackage, workingDirectory string, importPaths []string) (map[string]map[string]*Interface, error) {
	packageTypes := make(map[string]map[string]*Interface, len(importPaths))
	for _, importPath := range importPaths {
		path, dir := paths.ResolveImportPath(workingDirectory, importPath)
		log.Printf("parsing package '%s'\n", paths.GetRelativePath(dir))

		types, err := gatherTypesForPackage(pkgs, importPath, path)
		if err != nil {
			return nil, err
		}

		packageTypes[path] = types
	}

	return packageTypes, nil
}

func gatherTypesForPackage(pkgs []*internal.GoPackage, importPath, path string) (map[string]*Interface, error) {
	for _, pkg := range pkgs {
		if pkg.PkgPath != path {
			continue
		}

		for _, err := range pkg.Errors {
			switch err.Kind {
			case packages.TypeError:
				log.Printf("package %s failed to type check, but attempting to continue: %s", importPath, err.Msg)
			default:
				return nil, fmt.Errorf("malformed package %s (%s)", importPath, err.Msg)
			}
		}

		visitor := newVisitor(path, pkg.Types)
		for _, file := range pkg.Syntax {
			ast.Walk(visitor, file)
		}

		return visitor.types, nil
	}

	return nil, fmt.Errorf("malformed package %s (not found)", importPath)
}

func gatherAllPackageTypeNames(packageTypes map[string]map[string]*Interface) []string {
	nameMap := map[string]struct{}{}
	for _, pkg := range packageTypes {
		for name := range pkg {
			nameMap[name] = struct{}{}
		}
	}

	names := make([]string, 0, len(nameMap))
	for name := range nameMap {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}

func extractInterface(packageTypes map[string]map[string]*Interface, name string, targetNames, excludeNames []string) (*Interface, error) {
	if !shouldInclude(name, targetNames, excludeNames) {
		return nil, nil
	}

	candidates := make([]*Interface, 0, 1)
	for _, pkg := range packageTypes {
		if t, ok := pkg[name]; ok {
			candidates = append(candidates, t)

			if len(candidates) > 1 {
				return nil, fmt.Errorf("type '%s' is multiply-defined in supplied import paths", name)
			}
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	iface := candidates[0]

	for _, method := range iface.Methods {
		if !unicode.IsUpper([]rune(method.Name)[0]) {
			return nil, fmt.Errorf(
				"type '%s' has unexported an method '%s'",
				name,
				method.Name,
			)
		}
	}

	return iface, nil
}

func shouldInclude(name string, targetNames, excludeNames []string) bool {
	for _, v := range excludeNames {
		if strings.EqualFold(v, name) {
			return false
		}
	}

	for _, v := range targetNames {
		if strings.EqualFold(v, name) {
			return true
		}
	}

	return len(targetNames) == 0
}
