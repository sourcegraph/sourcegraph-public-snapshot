package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/loader"

	"github.com/kr/fs"

	"sourcegraph.com/sourcegraph/go-flags"
)

func init() {
	cliInit = append(cliInit, func(cli *flags.Command) {
		_, err := cli.AddCommand("validate",
			"use a simple heuristic to check that srclib is outputting expected graph data",
			`The validate command acts as a sanity check to ensure that a toolchain succeeds when it should and fails when it should not.`,
			&validateCmd,
		)
		if err != nil {
			log.Fatal(err)
		}
	})
}

var cacheDir = ".srclib-cache"
var unitFile = "GoPackage.unit.json"
var depFile = "GoPackage.depresolve.json"
var graphFile = "GoPackage.graph.json"

type ValidateCmd struct {
	w io.Writer
}

type Validate struct {
	Warnings []BuildWarning
}

type BuildWarning struct {
	Directory string
	Warning   string
}

const (
	BuildSucceededSrclibFailed = "Build succeeded but Srclib outputs failed"
	BuildFailedSrclibSucceeded = "Build failed but Srclib succeeded"
)

var validateCmd ValidateCmd

// Execute performs a sanity check on srclib builds for go repositories.
// We iterate every directory and do a validate check on every package.
// The validate heuristic is very rough currently, but makes the following assumptions:
// - only golang validate is checked
// - if a directory can be imported (see build standard library package) and built
//   (see loader standard library package), then there should be three files present
//   in the corresponding directory under .srclib-cache: a unit file, a depresolve file,
//   and a graph file.
func (c *ValidateCmd) Execute(args []string) error {
	if c.w == nil {
		c.w = os.Stdout
	}

	lRepo, lRepoErr := OpenLocalRepo()
	if lRepoErr != nil {
		return lRepoErr
	}

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		return errors.New("GOPATH not set")
	}

	var importPath string
	splitGoPath := filepath.SplitList(goPath)

	for _, p := range splitGoPath {
		p = filepath.ToSlash(p)
		if strings.Contains(lRepo.RootDir, p) {
			importPath = strings.TrimPrefix(lRepo.RootDir, filepath.Join(p, "src")+"/")
		}
	}

	if importPath == "" {
		return fmt.Errorf("Unable to find an import path for the current repo at %s", lRepo.RootDir)

	}

	cov := &Validate{}

	_, err := os.Stat(filepath.Join(lRepo.RootDir, cacheDir))
	if os.IsNotExist(err) {
		return err
	}

	walker := fs.Walk(lRepo.RootDir)

walkLoop:
	for walker.Step() {

		if err := walker.Err(); err != nil {
			return err
		}

		pth := walker.Path()

		fi, err := os.Stat(pth)
		if err != nil {
			return err
		} else if !fi.IsDir() {
			continue
		}

		for _, part := range strings.Split(pth, string(os.PathSeparator)) {
			if strings.HasPrefix(part, ".") {
				continue walkLoop
			}
		}

		relPath, err := filepath.Rel(lRepo.RootDir, pth)
		if err != nil {
			return err
		}

		_, importErr := build.ImportDir(pth, 0)

		var conf loader.Config
		conf.Import(strings.Join([]string{importPath, relPath}, "/"))
		_, pkgErr := conf.Load()

		importAndBuildSucceeded := (importErr == nil) && (pkgErr == nil)

		// TODO(poler) allow the user to specify an older commit (fine for now)
		cachePath := filepath.Join(lRepo.RootDir, cacheDir, lRepo.CommitID, importPath, relPath)

		// If the srclib build config was at all customized, the assumption that these files
		// will exist is almost certainly not valid.
		unitPath := filepath.Join(cachePath, unitFile)
		depPath := filepath.Join(cachePath, depFile)
		graphPath := filepath.Join(cachePath, graphFile)

		_, unitErr := os.Stat(unitPath)
		_, depErr := os.Stat(depPath)
		_, graphErr := os.Stat(graphPath)

		srclibOutputsExist := (unitErr == nil) && (depErr == nil) && (graphErr == nil)

		pth = filepath.ToSlash(pth)

		if importAndBuildSucceeded && !srclibOutputsExist {
			cov.Warnings = append(cov.Warnings, BuildWarning{
				Directory: pth,
				Warning:   BuildSucceededSrclibFailed,
			})
		} else if !importAndBuildSucceeded && srclibOutputsExist {
			cov.Warnings = append(cov.Warnings, BuildWarning{
				Directory: pth,
				Warning:   BuildFailedSrclibSucceeded,
			})
		}
	}

	enc := json.NewEncoder(c.w)
	if err := enc.Encode(cov); err != nil {
		return err
	}

	return nil

}
