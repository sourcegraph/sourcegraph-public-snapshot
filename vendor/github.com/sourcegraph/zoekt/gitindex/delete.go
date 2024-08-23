package gitindex

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// DeleteRepos deletes stale repos under a specific path in disk. The `names`
// argument stores names of repos retrieved from the git hosting site
// and is used along with the `filter` argument to decide on repo deletion.
func DeleteRepos(baseDir string, urlPrefix *url.URL, names map[string]struct{}, filter *Filter) error {
	paths, err := ListRepos(baseDir, urlPrefix)
	if err != nil {
		return err
	}
	var toDelete []string
	for _, p := range paths {
		_, exists := names[p]
		repoName := strings.Replace(p, filepath.Join(urlPrefix.Host, urlPrefix.Path), "", 1)
		repoName = strings.TrimPrefix(repoName, "/")
		if filter.Include(repoName) && !exists {
			toDelete = append(toDelete, p)
		}
	}

	if len(toDelete) > 0 {
		log.Printf("deleting repos %v", toDelete)
	}

	var errs []string
	for _, d := range toDelete {
		if err := os.RemoveAll(filepath.Join(baseDir, d)); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors: %v", errs)
	}
	return nil
}
