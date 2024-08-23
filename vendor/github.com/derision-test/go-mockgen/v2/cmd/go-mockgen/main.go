package main

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/derision-test/go-mockgen/v2/internal"
	"github.com/derision-test/go-mockgen/v2/internal/mockgen/generation"
	"github.com/derision-test/go-mockgen/v2/internal/mockgen/types"
	"golang.org/x/tools/go/packages"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("go-mockgen: ")
}

func main() {
	if err := mainErr(); err != nil {
		message := fmt.Sprintf("error: %s\n", err.Error())

		if solvableError, ok := err.(solvableError); ok {
			message += "\nPossible solutions:\n"

			for _, hint := range solvableError.Solutions() {
				message += fmt.Sprintf("  - %s\n", hint)
			}

			message += "\n"
		}

		log.Fatalf(message)
	}
}

type solvableError interface {
	Solutions() []string
}

func mainErr() error {
	allOptions, err := parseAndValidateOptions()
	if err != nil {
		return err
	}

	var (
		importPaths []string
		archives    []archive
		// map of import path to importpathToSourcefiles
		importpathToSourcefiles = make(map[string][]string)
	)

	var stdlibRoot string
	for _, opts := range allOptions {
		for _, packageOpts := range opts.PackageOptions {
			importPaths = append(importPaths, packageOpts.ImportPaths...)

			// this should be equal for opts.PackageOptions
			stdlibRoot = packageOpts.StdlibRoot

			importpath := packageOpts.ImportPaths[0]
			if len(packageOpts.SourceFiles) > 0 {
				importpathToSourcefiles[importpath] = append(importpathToSourcefiles[importpath], packageOpts.SourceFiles...)
			}

			for _, archive := range packageOpts.Archives {
				a, err := parseArchive(archive)
				if err != nil {
					return fmt.Errorf("error parsing achive mapping %q: %w", archive, err)
				}
				archives = append(archives, a)
			}
		}
	}

	// If multiple Sources reference the same importpath, we'll have n>1 copies of that importpath's
	// source files, which will cause "x redeclared in this block" errors.
	for importPath, sources := range importpathToSourcefiles {
		slices.Sort(sources)
		importpathToSourcefiles[importPath] = slices.Compact(sources)
	}

	log.Printf("loading data for %d packages\n", len(importPaths))

	pkgs, err := loadPackages(loadParams{
		importPaths: importPaths,
		// gcexportdata
		archives:   archives,
		sources:    importpathToSourcefiles,
		stdlibRoot: stdlibRoot,
	})
	if err != nil {
		return fmt.Errorf("could not load packages %s (%s)", strings.Join(importPaths, ","), err.Error())
	}

	for _, opts := range allOptions {
		typePackageOpts := make([]types.PackageOptions, 0, len(opts.PackageOptions))
		for _, packageOpts := range opts.PackageOptions {
			typePackageOpts = append(typePackageOpts, types.PackageOptions(packageOpts))
		}

		ifaces, err := types.Extract(pkgs, typePackageOpts)
		if err != nil {
			return err
		}

		nameMap := make(map[string]struct{}, len(ifaces))
		for _, t := range ifaces {
			nameMap[strings.ToLower(t.Name)] = struct{}{}
		}

		for _, packageOpts := range opts.PackageOptions {
			for _, name := range packageOpts.Interfaces {
				if _, ok := nameMap[strings.ToLower(name)]; !ok {
					return fmt.Errorf("type '%s' not found in supplied import paths", name)
				}
			}
		}

		if err := generation.Generate(ifaces, opts); err != nil {
			return err
		}
	}

	return nil
}

type loadParams struct {
	importPaths []string

	// gcexportdata specific params
	archives   []archive
	sources    map[string][]string
	stdlibRoot string
}

func loadPackages(params loadParams) ([]*internal.GoPackage, error) {
	if len(params.archives) > 0 || params.stdlibRoot != "" || len(params.sources) > 0 {
		packages, err := PackagesArchive(params)
		if err != nil {
			return nil, fmt.Errorf("error loading packages from archives: %v", err)
		}
		return packages, nil
	}

	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName | packages.NeedImports | packages.NeedSyntax | packages.NeedTypes | packages.NeedDeps}, params.importPaths...)
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, errors.New("no packages found")
	}

	ipkgs := make([]*internal.GoPackage, 0, len(pkgs))
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			var errString string
			for _, err := range pkg.Errors {
				errString += err.Error() + "\n"
			}
			errString = strings.TrimSpace(errString)
			return nil, errors.New(errString)
		}
		ipkgs = append(ipkgs, internal.NewPackage(pkg))
	}
	return ipkgs, nil
}
