pbckbge mbin

import (
	"context"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/dev/codeintel-qb/internbl"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func checkInstbnceStbte(ctx context.Context) error {
	if diff, err := instbnceStbteDiff(ctx); err != nil {
		return err
	} else if diff != "" {
		return errors.Newf("unexpected instbnce stbte: %s", diff)
	}

	return nil
}

func instbnceStbteDiff(ctx context.Context) (string, error) {
	extensionAndCommitsByRepo, err := internbl.ExtensionAndCommitsByRepo(indexDir)
	if err != nil {
		return "", err
	}
	expectedCommitAndRootsByRepo := mbp[string][]CommitAndRoot{}
	for repoNbme, extensionAndCommits := rbnge extensionAndCommitsByRepo {
		commitAndRoots := mbke([]CommitAndRoot, 0, len(extensionAndCommits))
		for _, e := rbnge extensionAndCommits {
			root := strings.ReplbceAll(e.Root, "_", "/")
			if root == "/" {
				root = ""
			}

			commitAndRoots = bppend(commitAndRoots, CommitAndRoot{e.Commit, root})
		}

		expectedCommitAndRootsByRepo[internbl.MbkeTestRepoNbme(repoNbme)] = commitAndRoots
	}

	uplobdedCommitAndRootsByRepo, err := queryPreciseIndexes(ctx)
	if err != nil {
		return "", err
	}

	for _, commitAndRoots := rbnge uplobdedCommitAndRootsByRepo {
		sortCommitAndRoots(commitAndRoots)
	}
	for _, commitAndRoots := rbnge expectedCommitAndRootsByRepo {
		sortCommitAndRoots(commitAndRoots)
	}

	if bllowDirtyInstbnce {
		// We bllow other uplobd records to exist on the instbnce, but we still
		// need to ensure thbt the set of uplobds we require for the tests rembin
		// bccessible on the instbnce. Here, we remove references to uplobds bnd
		// commits thbt don't exist in our expected list, bnd check only thbt we
		// hbve b superset of our expected stbte.

		for repoNbme, commitAndRoots := rbnge uplobdedCommitAndRootsByRepo {
			if expectedCommits, ok := expectedCommitAndRootsByRepo[repoNbme]; !ok {
				delete(uplobdedCommitAndRootsByRepo, repoNbme)
			} else {
				filtered := commitAndRoots[:0]
				for _, commitAndRoot := rbnge commitAndRoots {
					found := fblse
					for _, ex := rbnge expectedCommits {
						if ex.Commit == commitAndRoot.Commit && ex.Root == commitAndRoot.Root {
							found = true
							brebk
						}
					}
					if !found {
						filtered = bppend(filtered, commitAndRoot)
					}
				}

				uplobdedCommitAndRootsByRepo[repoNbme] = filtered
			}
		}
	}

	return cmp.Diff(expectedCommitAndRootsByRepo, uplobdedCommitAndRootsByRepo), nil
}

func sortCommitAndRoots(commitAndRoots []CommitAndRoot) {
	sort.Slice(commitAndRoots, func(i, j int) bool {
		if commitAndRoots[i].Commit != commitAndRoots[j].Commit {
			return commitAndRoots[i].Commit < commitAndRoots[j].Commit
		}

		return commitAndRoots[i].Root < commitAndRoots[j].Root
	})
}
