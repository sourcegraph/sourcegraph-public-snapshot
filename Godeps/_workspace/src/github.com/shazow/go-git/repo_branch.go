package git

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
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
	return repo.listRefs("refs/heads/")
}

func (repo *Repository) CreateBranch(branchName, idStr string) error {
	return repo.createRef("heads", branchName, idStr)
}

func (repo *Repository) createRef(head, branchName, idStr string) error {
	id, err := NewIdFromString(idStr)
	if err != nil {
		return err
	}

	branchPath := filepath.Join(repo.Path, "refs/"+head, branchName)
	if isFile(branchPath) {
		return ErrBranchExisted
	}

	f, err := os.Create(branchPath)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.WriteString(f, id.String())
	return err
}

// listRefs returns a list of refs that begin with prefix. The prefix
// has been trimmed from the refs that are returned (e.g., "mybranch"
// not "refs/heads/mybranch").
func (repo *Repository) listRefs(prefix string) ([]string, error) {
	var refs []string

	dir := filepath.Join(repo.Path, prefix)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == ".DS_Store" {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		ref, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		refs = append(refs, ref)
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Refs can also be packed refs.
	packedRefs, err := repo.listPackedRefs(prefix)
	if err != nil {
		return nil, err
	}
	refs = append(refs, packedRefs...)
	return refs, nil
}

func (repo *Repository) listPackedRefs(prefix string) (refs []string, err error) {
	packedRefs, err := ioutil.ReadFile(filepath.Join(repo.Path, "packed-refs"))
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	for _, line := range bytes.Split(packedRefs, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || bytes.HasPrefix(line, []byte("#")) {
			continue
		}

		// Line format example: "ea167fe3d76b1e5fd3ed8ca44cbd2fe3897684f8 refs/heads/b0"
		parts := bytes.SplitN(line, []byte(" "), 2)
		if len(parts) != 2 {
			continue
		}
		ref := string(parts[1])
		if strings.HasPrefix(ref, prefix) {
			refs = append(refs, strings.TrimPrefix(ref, prefix))
		}
	}
	return
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
