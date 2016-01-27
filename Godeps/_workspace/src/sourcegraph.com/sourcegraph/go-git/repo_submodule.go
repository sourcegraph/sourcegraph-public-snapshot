package git

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

// GetModules returns the submodule names that are initialized and cached within .git/modules
func (repo *Repository) GetModules() ([]string, error) {
	dirPath := filepath.Join(repo.Path, "modules")
	fis, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(fis))
	for _, fi := range fis {
		if strings.Contains(fi.Name(), ".DS_Store") {
			continue
		}
		if !fi.IsDir() {
			continue
		}

		names = append(names, fi.Name())
	}

	return names, nil
}
