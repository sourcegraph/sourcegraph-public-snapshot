package idx

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// index updates the cross-repo code intelligence indexes for the given repository at the given revision and enqueues
// further repos to index if applicable.
func (w *Worker) index(repoName api.RepoURI, rev string, isPrimary bool) (err error) {
	repo, commit, err := resolveRevision(w.Ctx, repoName, rev)
	if err != nil {
		// Avoid infinite loop for always cloning test.
		if repo != nil && repo.URI == "github.com/sourcegraphtest/AlwaysCloningTest" {
			return nil
		}
		return err
	}

	// Check if index is already up-to-date
	if repo.IndexedRevision != nil && (repo.FreezeIndexedRevision || *repo.IndexedRevision == commit) {
		return nil
	}

	// Get language
	inv, err := api.InternalClient.ReposGetInventoryUncached(w.Ctx, repo.ID, commit)
	if err != nil {
		return fmt.Errorf("Repos.GetInventory failed: %s", err)
	}
	lang := inv.PrimaryProgrammingLanguage()

	// Update global refs & packages index
	if !repo.Fork() {
		var errs []error
		if err := api.InternalClient.DefsRefreshIndex(w.Ctx, repo.URI, commit); err != nil {
			errs = append(errs, fmt.Errorf("Defs.RefreshIndex failed: %s", err))
		}

		if err := api.InternalClient.PkgsRefreshIndex(w.Ctx, repo.URI, commit); err != nil {
			errs = append(errs, fmt.Errorf("Pkgs.RefreshIndex failed: %s", err))
		}

		if err := makeMultiErr(errs...); err != nil {
			return err
		}
	}

	err = api.InternalClient.ReposUpdateIndex(w.Ctx, repo.ID, commit, lang)
	if err != nil {
		return err
	}
	return nil
}

type multiError []error

func makeMultiErr(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}
	var nonnil []error
	for _, err := range errs {
		if err != nil {
			nonnil = append(nonnil, err)
		}
	}
	if len(nonnil) == 0 {
		return nil
	}
	if len(nonnil) == 1 {
		return nonnil[0]
	}
	return multiError(nonnil)
}

func (e multiError) Error() string {
	errs := make([]string, len(e))
	for i := 0; i < len(e); i++ {
		errs[i] = e[i].Error()
	}
	return fmt.Sprintf("multiple errors:\n\t%s", strings.Join(errs, "\n\t"))
}
