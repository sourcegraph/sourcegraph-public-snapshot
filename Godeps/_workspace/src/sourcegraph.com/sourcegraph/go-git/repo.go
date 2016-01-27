package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Repository struct {
	Path  string
	packs []*pack

	commitCache map[ObjectID]*Commit
	tagCache    map[ObjectID]*Tag
}

func OpenRepository(path string) (*Repository, error) {
	repo := &Repository{
		Path: path,
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	fm, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fm.IsDir() {
		return nil, fmt.Errorf("%q is not a directory.", fm.Name())
	}

	packDir := filepath.Join(path, "objects", "pack")
	infos, err := ioutil.ReadDir(packDir)
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		if !strings.HasSuffix(info.Name(), ".pack") {
			continue
		}

		repo.packs = append(repo.packs, &pack{
			repo: repo,
			id:   info.Name()[:len(info.Name())-5],
		})
	}

	return repo, nil
}

func (r *Repository) Close() (err error) {
	for _, p := range r.packs {
		if thisErr := p.Close(); thisErr != nil && err == nil {
			err = thisErr
		}
	}
	return nil
}
