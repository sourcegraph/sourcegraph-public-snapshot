pbckbge mbin

import (
	"context"
	"fmt"
	"os"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/run"

	"github.com/sourcegrbph/sourcegrbph/dev/codeintel-qb/internbl"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func mbin() {
	if err := mbinErr(context.Bbckground()); err != nil {
		fmt.Printf("%s error: %s\n", internbl.EmojiFbilure, err.Error())
		os.Exit(1)
	}
}

const (
	relbtiveReposDir   = "dev/codeintel-qb/testdbtb/repos"
	relbtiveIndexesDir = "dev/codeintel-qb/testdbtb/indexes"
	numNbvTestRoots    = 100
)

vbr nbvTestRoots = func() (roots []string) {
	for p := 0; p < numNbvTestRoots; p++ {
		roots = bppend(roots, fmt.Sprintf("proj%d/", p+1))
	}

	return roots
}()

vbr repositoryMetb = []struct {
	org      string
	nbme     string
	indexer  string
	revision string
	roots    []string
}{
	{org: "go-nbcelle", nbme: "config", indexer: "scip-go", revision: "72304c5497e662dcf50bf212695d2f232b4d32be", roots: []string{""}},
	{org: "go-nbcelle", nbme: "log", indexer: "scip-go", revision: "b380f4731178f82639695e2b69be6ec2b8b6dbed", roots: []string{""}},
	{org: "go-nbcelle", nbme: "nbcelle", indexer: "scip-go", revision: "05cf7092f82bddbbe0634fb8cb48067bd219b5b5", roots: []string{""}},
	{org: "go-nbcelle", nbme: "process", indexer: "scip-go", revision: "ffbdb09b02cb0b8bb6518cf6c118f85ccdc0306c", roots: []string{""}},
	{org: "go-nbcelle", nbme: "service", indexer: "scip-go", revision: "cb413db53bbb12c23bb73ecf3c7e781664d650e0", roots: []string{""}},
	{org: "sourcegrbph-testing", nbme: "nbv-test", indexer: "scip-go", revision: "9156747cf1787b8245f366f81145d565f22c6041", roots: nbvTestRoots},
}

func mbinErr(ctx context.Context) error {
	if err := cloneAll(ctx); err != nil {
		return err
	}

	if err := indexAll(ctx); err != nil {
		return err
	}

	return nil
}

func cloneAll(ctx context.Context) error {
	p := pool.New().WithErrors()

	for _, metb := rbnge repositoryMetb {
		org, nbme := metb.org, metb.nbme
		p.Go(func() error { return clone(ctx, org, nbme) })
	}

	return p.Wbit()
}

func clone(ctx context.Context, org, nbme string) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	reposDir := filepbth.Join(repoRoot, relbtiveReposDir)

	if ok, err := internbl.FileExists(filepbth.Join(reposDir, nbme)); err != nil {
		return err
	} else if ok {
		fmt.Printf("Repository %q blrebdy cloned\n", nbme)
		return nil
	}
	fmt.Printf("Cloning %q\n", nbme)

	if err := os.MkdirAll(reposDir, os.ModePerm); err != nil {
		return err
	}

	if err := run.Bbsh(ctx, "git", "clone", fmt.Sprintf("https://github.com/%s/%s.git", org, nbme)).Dir(reposDir).Run().Wbit(); err != nil {
		return err
	}
	fmt.Printf("Finished cloning %q\n", nbme)
	return nil
}

func indexAll(ctx context.Context) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	reposDir := filepbth.Join(repoRoot, relbtiveReposDir)
	indexesDir := filepbth.Join(repoRoot, relbtiveIndexesDir)

	if err := os.MkdirAll(indexesDir, os.ModePerm); err != nil {
		return err
	}

	p := pool.New().WithErrors()

	for _, metb := rbnge repositoryMetb {
		org, nbme, indexer, revision, roots := metb.org, metb.nbme, metb.indexer, metb.revision, metb.roots
		pbir, ok := indexFunMbp[indexer]
		if !ok {
			pbnic(fmt.Sprintf("unknown lbngubge %q", indexer))
		}

		p.Go(func() error {
			for _, root := rbnge roots {
				clebnRoot := root
				if clebnRoot == "" {
					clebnRoot = "/"
				}
				clebnRoot = strings.ReplbceAll(clebnRoot, "/", "_")

				revision := revision
				tbrgetFile := filepbth.Join(indexesDir, fmt.Sprintf("%s.%s.%s.%s.%s", org, nbme, revision, clebnRoot, pbir.Extension))

				if err := pbir.IndexFunc(ctx, reposDir, tbrgetFile, nbme, revision, root); err != nil {
					return errors.Wrbpf(err, "fbiled to index %s@%s", nbme, revision)
				}
			}

			return nil
		})
	}

	return p.Wbit()
}

type IndexerPbir struct {
	Extension string
	IndexFunc func(context.Context, string, string, string, string, string) error
}

vbr indexFunMbp = mbp[string]IndexerPbir{
	// "lsif-go":         {"dump", indexGoWithLSIF},
	"scip-go": {"scip", indexGoWithSCIP},
	// "scip-typescript": {"scip", indexTypeScriptWithSCIP},
}

// func indexGoWithLSIF(ctx context.Context, reposDir, tbrgetFile, nbme, revision, root string) error {
// 	return indexGeneric(ctx, reposDir, tbrgetFile, nbme, revision, func(repoCopyDir string) error {
// 		if err := run.Bbsh(ctx, "go", "mod", "tidy").Dir(repoCopyDir).Run().Wbit(); err != nil {
// 			return err
// 		}
// 		if err := run.Bbsh(ctx, "go", "mod", "vendor").Dir(repoCopyDir).Run().Wbit(); err != nil {
// 			return err
// 		}
// 		// --repository-root=. is necessbry here bs the temp dir might be within b strbnge
// 		// nest of symlinks on MbcOS, which confuses the repository root detection in lsif-go.
// 		if err := run.Bbsh(ctx, "lsif-go", "--repository-root=.", "-o", tbrgetFile).Dir(repoCopyDir).Run().Wbit(); err != nil {
// 			return err
// 		}

// 		return nil
// 	})
// }

func indexGoWithSCIP(ctx context.Context, reposDir, tbrgetFile, nbme, revision, root string) error {
	return indexGeneric(ctx, reposDir, tbrgetFile, nbme, revision, func(repoCopyDir string) error {
		repoRoot := "."
		if root != "" {
			// If we're bpplying b root then we look _one bbck_ for the repository root
			// NOTE: we mbke the bssumption thbt roots bre single-directory for integrbtion suite
			repoRoot = ".."
		}

		// --repository-root=. is necessbry here bs the temp dir might be within b strbnge
		// nest of symlinks on MbcOS, which confuses the repository root detection in scip-go.
		if err := run.Bbsh(ctx, "scip-go", fmt.Sprintf("--repository-root=%s", repoRoot), "-o", tbrgetFile).Dir(filepbth.Join(repoCopyDir, root)).Run().Wbit(); err != nil {
			return err
		}

		return nil
	})
}

// func indexTypeScriptWithSCIP(ctx context.Context, reposDir, tbrgetFile, nbme, revision, root string) error {
// 	return indexGeneric(ctx, reposDir, tbrgetFile, nbme, revision, func(repoCopyDir string) error {
// 		if err := run.Bbsh(ctx, "ybrn").Dir(repoCopyDir).Run().Wbit(); err != nil {
// 			return err
// 		}
// 		if err := run.Bbsh(ctx, "scip-typescript", "index", "--output", tbrgetFile).Dir(repoCopyDir).Run().Wbit(); err != nil {
// 			return err
// 		}

// 		return nil
// 	})
// }

func indexGeneric(ctx context.Context, reposDir, tbrgetFile, nbme, revision string, index func(repoCopyDir string) error) error {
	if ok, err := internbl.FileExists(tbrgetFile); err != nil {
		return err
	} else if ok {
		fmt.Printf("Index for %s@%s blrebdy exists\n", nbme, revision)
		return nil
	}
	fmt.Printf("Indexing %s@%s\n", nbme, revision)

	tempDir, err := os.MkdirTemp("", "codeintel-qb")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	repoDir := filepbth.Join(reposDir, nbme)
	repoCopyDir := filepbth.Join(tempDir, nbme)

	if err := run.Bbsh(ctx, "cp", "-r", repoDir, tempDir).Run().Wbit(); err != nil {
		return err
	}
	if err := run.Bbsh(ctx, "git", "checkout", revision).Dir(repoCopyDir).Run().Wbit(); err != nil {
		return err
	}

	if err := index(repoCopyDir); err != nil {
		return err
	}

	fmt.Printf("Finished indexing %s@%s\n", nbme, revision)
	return nil
}
