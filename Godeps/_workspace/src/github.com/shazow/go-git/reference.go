package git

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// UnpackRefs unpacks 'packed-refs' to git repository.
func UnpackRefs(repoPath string) error {
	refs, err := ioutil.ReadFile(filepath.Join(repoPath, "packed-refs"))
	if err != nil {
		return err
	}

	for _, ref := range strings.Split(string(refs), "\n") {
		if len(ref) == 0 || ref[0] == '#' {
			continue
		} else if !strings.Contains(ref, "heads") && !strings.Contains(ref, "tags") {
			continue
		}

		infos := strings.Split(ref, " ")
		refPath := filepath.Join(repoPath, infos[1])
		os.RemoveAll(refPath)
		os.MkdirAll(path.Dir(refPath), os.ModePerm)
		if err = ioutil.WriteFile(refPath, []byte(infos[0]), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
