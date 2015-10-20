package gendata

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/dep"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// hierarchicalNames returns a slice of hierarchical filenames with branching factor at each
// level specified by structure
func hierarchicalNames(nodeRoot string, leafRoot string, prefix string, structure []int) (filenames []string) {
	if len(structure) == 0 {
		return nil
	}
	if len(structure) == 1 {
		nfiles := structure[0]
		for i := 0; i < nfiles; i++ {
			filenames = append(filenames, filepath.Join(prefix, fmt.Sprintf("%s_%d", leafRoot, i)))
		}
		return filenames
	}

	head, tail := structure[0], structure[1:]
	for i := 0; i < head; i++ {
		subdir := filepath.Join(prefix, fmt.Sprintf("%s_%d", nodeRoot, i))
		filenames = append(filenames, hierarchicalNames(nodeRoot, leafRoot, subdir, tail)...)
	}
	return filenames
}

func getGitCommitID() (string, error) {
	err := exec.Command("git", "add", "-A", ":/").Run()
	if err != nil {
		return "", err
	}
	err = exec.Command("git", "commit", "-m", "generated source").Run()
	if err != nil {
		return "", err
	}
	out, err := exec.Command("git", "log", "--pretty=oneline", "-n1").Output()
	if err != nil {
		return "", err
	}
	commitID := strings.Fields(string(out))[0]
	return commitID, nil
}

func writeSrclibCache(ut *unit.SourceUnit, gr *graph.Output, deps []*dep.Resolution) error {
	unitDir := filepath.Join(".srclib-cache", ut.CommitID, ut.Name)
	if err := os.MkdirAll(unitDir, 0700); err != nil {
		return err
	}

	unitFile, err := os.OpenFile(filepath.Join(unitDir, fmt.Sprintf("%s.unit.json", ut.Type)), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer unitFile.Close()

	if err := json.NewEncoder(unitFile).Encode(ut); err != nil {
		return err
	}

	graphFile, err := os.OpenFile(filepath.Join(unitDir, fmt.Sprintf("%s.graph.json", ut.Type)), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer graphFile.Close()

	if err := json.NewEncoder(graphFile).Encode(gr); err != nil {
		return err
	}

	depFile, err := os.OpenFile(filepath.Join(unitDir, fmt.Sprintf("%s.depresolve.json", ut.Type)), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer depFile.Close()

	if err := json.NewEncoder(depFile).Encode(deps); err != nil {
		return err
	}
	return nil
}

func resetSource() error {
	if err := removeGlob("unit_*"); err != nil {
		return err
	}
	if err := removeGlob("u_*"); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		clearReadOnly(".git")
	}
	if err := os.RemoveAll(".git"); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := exec.Command("git", "init").Run(); err != nil {
		return err
	}
	return nil
}

func removeGlob(glob string) error {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	for _, match := range matches {
		if err := os.RemoveAll(match); err != nil {
			return err
		}
	}
	return nil
}

// Tries to remove READONLY mark from file (recursive)
// On Windows, os.Remove does not work if file was marked as READONLY (for example, git does it)
func clearReadOnly(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if !fi.IsDir() {
		return os.Chmod(path, 0666)
	}

	fd, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fd.Close()

	names, _ := fd.Readdirnames(-1)
	for _, name := range names {
		err = clearReadOnly(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}
	return nil
}
