package python

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kr/fs"

	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// Scan a directory, listing all source units
func Scan(srcdir string, repoURI string, repoSubdir string) ([]*unit.SourceUnit, error) {
	if units, isSpecial := specialUnits[repoURI]; isSpecial {
		return units, nil
	}

	p, err := getVENVBinPath()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(filepath.Join(p, "pydep-run.py"), "list", srcdir)

	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var pkgs []*pkgInfo
	if err := json.NewDecoder(stdout).Decode(&pkgs); err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	// Keep track of all files that have been successfully discovered
	discoveredScripts := make(map[string]bool)

	units := make([]*unit.SourceUnit, len(pkgs))
	for i, pkg := range pkgs {
		units[i] = pkg.SourceUnit()
		units[i].Files = pythonSourceFiles(pkg.RootDir, discoveredScripts)
		units[i].Repo = repoURI // override whatever's in the setup.py file with the actual reposit;5Dory URI

		reqs, err := requirements(pkg.RootDir)
		if err != nil {
			return nil, err
		}
		reqs_ := make([]interface{}, len(reqs))
		for i, req := range reqs {
			reqs_[i] = req
		}
		units[i].Dependencies = reqs_
	}

	// Scan for independant scripts, appending to the current set of source units
	scripts := pythonSourceFiles(srcdir, discoveredScripts)
	if len(scripts) > 0 {
		scriptsUnit := unit.SourceUnit{
			Name:  ".",
			Type:  "PythonProgram",
			Files: scripts,
			Dir:   ".",
			Ops:   map[string]*srclib.ToolRef{"depresolve": nil, "graph": nil},
		}

		reqs, err := requirements(srcdir)
		if err == nil {
			reqs_ := make([]interface{}, len(reqs))
			for i, req := range reqs {
				reqs_[i] = req
			}
			scriptsUnit.Dependencies = reqs_
		}

		units = append(units, &scriptsUnit)
	}

	return units, nil
}

func requirements(unitDir string) ([]*requirement, error) {
	p, err := getVENVBinPath()
	if err != nil {
		return nil, err
	}
	depCmd := exec.Command(filepath.Join(p, "pydep-run.py"), "dep", unitDir)
	depCmd.Stderr = os.Stderr
	b, err := depCmd.Output()
	if err != nil {
		return nil, err
	}

	var reqs []*requirement
	err = json.Unmarshal(b, &reqs)
	if err != nil {
		return nil, err
	}
	reqs, ignoredReqs := pruneReqs(reqs)
	if len(ignoredReqs) > 0 {
		ignoredKeys := make([]string, len(ignoredReqs))
		for r, req := range ignoredReqs {
			ignoredKeys[r] = req.Key
		}
		log.Printf("(warn) ignoring dependencies %v because repo URL absent", ignoredKeys)
	}

	return reqs, nil
}

// Get all python source files under dir
func pythonSourceFiles(dir string, discoveredScripts map[string]bool) (files []string) {
	walker := fs.Walk(dir)
	for walker.Step() {
		if err := walker.Err(); err == nil && !walker.Stat().IsDir() && filepath.Ext(walker.Path()) == ".py" {
			file := walker.Path()
			_, found := discoveredScripts[file]

			if !found {
				files = append(files, file)
				discoveredScripts[file] = true
			}
		}
	}
	return
}

// Remove unresolvable requirements (i.e., requirements with no clone URL)
func pruneReqs(reqs []*requirement) (kept, ignored []*requirement) {
	for _, req := range reqs {
		if req.RepoURL != "" { // cannot resolve dependencies with no clone URL
			kept = append(kept, req)
		} else {
			ignored = append(ignored, req)
		}
	}
	return
}
