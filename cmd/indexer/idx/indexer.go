package idx

import (
	"fmt"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/inventory"
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
		w.handleLangServersMigration(repo, commit)
		return nil
	}

	// Get language
	inv, err := api.InternalClient.ReposGetInventoryUncached(w.Ctx, repo.ID, commit)
	if err != nil {
		return fmt.Errorf("Repos.GetInventory failed: %s", err)
	}
	lang := inv.PrimaryProgrammingLanguage()

	// Automatically enable the language server for each language detected in
	// the repository.
	w.enableLangservers(inv)

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

// enableLangservers enables language servers for the given repo inventory.
func (w *Worker) enableLangservers(inv *inventory.Inventory) {
	if conf.IsDeployTypeKubernetesCluster(conf.DeployType()) {
		// Deployed to a cluster, do not auto-enable language servers.
		return
	}
	if conf.IsDev(conf.DeployType()) && !conf.DebugManageDocker() {
		// Running in dev mode with managed docker disabled.
		return
	}

	// Enable the language server for each language detected in the repository.
	for _, language := range inv.Languages {
		err := langServersEnableLanguage(w.Ctx, language.ConfigName())

		// Note: We ignore "not authenticated" errors here as they would just
		// indicate one of three things:
		//
		// 1. The language is a built-in one, but an admin user has explicitly
		//    disabled it. We do not want to act as an admin and explicitly
		//    override their disable action.
		// 2. The language is built-in and experimental. Only admins can do
		//    this.
		// 3. The language is not a built-in one, and by 'enabling' it we would
		//    actually just be modifying an entry to the site config. Only
		//    admins can do this, and this is not an action we want to perform
		//    here.
		//
		if err != nil && !strings.Contains(err.Error(), "not authenticated") {
			// Failure here should never be fatal to the rest of the
			// indexing operations.
			log15.Error("failed to automatically enable language server", "language", language.ConfigName(), "error", err)
		}
	}
}

// handleLangServersMigration handles a migration path: The repository is
// already indexed. But it is still useful to enable language servers
// automatically at this point for users who've already added all of their
// repositories before we supported automatic language server management.
//
// TODO(slimsag): remove this code after May 3, 2018. Also remove
// api.InternalClient.ReposGetInventory since this is the only user of it.
func (w *Worker) handleLangServersMigration(repo *api.Repo, commit api.CommitID) {
	// Note: we use ReposGetInventory not the uncached variant because this
	// runs on every refresh operation and it needs to be performant.
	inv, err := api.InternalClient.ReposGetInventory(w.Ctx, repo.ID, commit)
	if err != nil {
		log15.Error("failed to automatically enable language servers, Repos.GetInventory failed", "error", err)
		return
	}

	// Automatically enable the language server for each language detected in
	// the repository.
	w.enableLangservers(inv)
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
