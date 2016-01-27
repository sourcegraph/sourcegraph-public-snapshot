package git

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrBranchExisted = errors.New("branch has existed")
)

func IsBranchExist(repoPath, branchName string) bool {
	branchPath := filepath.Join(repoPath, "refs/heads", branchName)
	return isFile(branchPath)
}

func (repo *Repository) IsBranchExist(branchName string) bool {
	branchPath := filepath.Join(repo.Path, "refs/heads", branchName)
	return isFile(branchPath)
}

func (repo *Repository) GetBranches() ([]string, error) {
	return repo.readRefDir("refs/heads", "")
}

func (repo *Repository) CreateBranch(branchName, idStr string) error {
	return repo.createRef("heads", branchName, idStr)
}

func (repo *Repository) createRef(head, branchName, idStr string) error {
	branchPath := filepath.Join(repo.Path, "refs/"+head, branchName)
	if isFile(branchPath) {
		return ErrBranchExisted
	}

	f, err := os.Create(branchPath)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.WriteString(f, idStr)
	return err
}

func (repo *Repository) readRefDir(prefix, relPath string) ([]string, error) {
	dirPath := filepath.Join(repo.Path, prefix, relPath)
	f, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fis, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(fis))
	for _, fi := range fis {
		if strings.Contains(fi.Name(), ".DS_Store") {
			continue
		}

		relFileName := filepath.Join(relPath, fi.Name())
		if fi.IsDir() {
			subnames, err := repo.readRefDir(prefix, relFileName)
			if err != nil {
				return nil, err
			}
			names = append(names, subnames...)
			continue
		}

		names = append(names, relFileName)
	}

	return names, nil
}

func CreateBranch(repoPath, branchName, id string) error {
	return CreateRef("heads", repoPath, branchName, id)
}

func CreateRef(head, repoPath, branchName, id string) error {
	branchPath := filepath.Join(repoPath, "refs/"+head, branchName)
	if isFile(branchPath) {
		return ErrBranchExisted
	}
	f, err := os.Create(branchPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.WriteString(f, id)
	return err
}
