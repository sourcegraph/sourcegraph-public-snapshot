package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/codeintel-qa/internal"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func main() {
	if err := mainErr(context.Background()); err != nil {
		fmt.Printf("%s error: %s\n", internal.EmojiFailure, err.Error())
		os.Exit(1)
	}
}

const (
	relativeReposDir   = "dev/codeintel-qa/testdata/repos"
	relativeIndexesDir = "dev/codeintel-qa/testdata/indexes"
)

var repositoryMeta = []struct {
	org       string
	name      string
	indexer   string
	revisions []string
}{
	// This repository has not been changed from its upstream
	{
		org:     "sourcegraph-testing",
		name:    "zap",
		indexer: "lsif-go",
		revisions: []string{
			"a6015e13fab9b744d96085308ce4e8f11bad1996",
			"2aa9fa25da83bdfff756c36a91442edc9a84576c",
		},
	},

	//  Each commit here is tagged as sg-test-1, sg-test-2, and sg-test-3, respectively. See CHANGES.md in the root of the
	//  repository's master branch to see a history of changes and which revisions were targeted. We specifically use replace
	//  directives in the project root's go.mod file to target sourcegraph-testing/zap, which has no changes of its own. This
	//  simulates how common forking works in the Go ecosystem (see our own use of zoekt).
	//
	//  To ensure that the last commit in the list for each repository is visible at tip, the master branch's last commit is
	//  a merge commit between the true upstream tip and sg-test-3.
	{
		org:     "sourcegraph-testing",
		name:    "etcd",
		indexer: "lsif-go",
		revisions: []string{
			"4397ceb9c11be0b3e9ee0111230235c868ba581d",
			"bc588b7a2e9af4f903396cdcf66f56190b9e254f",
			"ad7848014a051dbe3fcd6a4cff2c7befdd16d5a8",
		},
	},
	{
		org:     "sourcegraph-testing",
		name:    "tidb",
		indexer: "lsif-go",
		revisions: []string{
			"8eaaa098b4e938b18485f7b1fa7d8e720b04c699",
			"b5f100a179e20d5539e629bd0919d05774cb7c6a",
			"9aab49176993f9dc0ed2fcb9ef7e5125518e8b98",
		},
	},
	{
		org:     "sourcegraph-testing",
		name:    "titan",
		indexer: "lsif-go",
		revisions: []string{
			"fb38de395ba67f49978b218e099de1c45122fb50",
			"415ffd5a3ba7a92a07cd96c7d9f4b734f61248f7",
			"f8307e394c512b4263fc0cd67ccf9fd46f1ad9a5",
		},
	},

	// These repositories have their module names modified and new tags created to refer to each other
	{
		org:     "sourcegraph-testing",
		name:    "nacelle",
		indexer: "scip-go",
		revisions: []string{
			"68d3125fb03d4aec540714577401f9f01adffa8a",
		},
	},
	{
		org:     "sourcegraph-testing",
		name:    "nacelle-config",
		indexer: "scip-go",
		revisions: []string{
			"4d4864d3b5b046fe12154f3aae7a86a04690c4ae",
		},
	},
	{
		org:     "sourcegraph-testing",
		name:    "nacelle-service",
		indexer: "scip-go",
		revisions: []string{
			"0652f3023c1bc7e7466a487f20bbe4b5e28fdcc7",
		},
	},

	// This repository is archived in-practice and as a good candidate for a low-effort scip-typescript test
	{
		org:     "sourcegraph",
		name:    "code-intel-extensions",
		indexer: "scip-typescript",
		revisions: []string{
			"c66e756d3d68a1e19048c3f7515ba42a7e793767",
		},
	},
}

func mainErr(ctx context.Context) error {
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

	for _, meta := range repositoryMeta {
		org, name := meta.org, meta.name
		p.Go(func() error { return clone(ctx, org, name) })
	}

	return p.Wait()
}

func clone(ctx context.Context, org, name string) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	reposDir := filepath.Join(repoRoot, relativeReposDir)

	if ok, err := internal.FileExists(filepath.Join(reposDir, name)); err != nil {
		return err
	} else if ok {
		fmt.Printf("Repository %q already cloned\n", name)
		return nil
	}
	fmt.Printf("Cloning %q\n", name)

	if err := os.MkdirAll(reposDir, os.ModePerm); err != nil {
		return err
	}

	if err := run.Bash(ctx, "git", "clone", fmt.Sprintf("https://github.com/%s/%s.git", org, name)).Dir(reposDir).Run().Wait(); err != nil {
		return err
	}
	fmt.Printf("Finished cloning %q\n", name)
	return nil
}

func indexAll(ctx context.Context) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	reposDir := filepath.Join(repoRoot, relativeReposDir)
	indexesDir := filepath.Join(repoRoot, relativeIndexesDir)

	if err := os.MkdirAll(indexesDir, os.ModePerm); err != nil {
		return err
	}

	p := pool.New().WithErrors()

	for _, meta := range repositoryMeta {
		org, name, indexer, revisions := meta.org, meta.name, meta.indexer, meta.revisions
		pair, ok := indexFunMap[indexer]
		if !ok {
			panic(fmt.Sprintf("unknown language %q", indexer))
		}

		p.Go(func() error {
			for i, revision := range revisions {
				revision := revision
				targetFile := filepath.Join(indexesDir, fmt.Sprintf("%s.%s.%d.%s.%s", org, name, i, revision, pair.Extension))

				if err := pair.IndexFunc(ctx, reposDir, targetFile, name, revision); err != nil {
					return errors.Wrapf(err, "failed to index %s@%s", name, revision)
				}
			}

			return nil
		})
	}

	return p.Wait()
}

type IndexerPair struct {
	Extension string
	IndexFunc func(context.Context, string, string, string, string) error
}

var indexFunMap = map[string]IndexerPair{
	"lsif-go":         {"dump", indexGoWithLSIF},
	"scip-go":         {"scip", indexGoWithSCIP},
	"scip-typescript": {"scip", indexTypeScriptWithSCIP},
}

func indexGoWithLSIF(ctx context.Context, reposDir, targetFile, name, revision string) error {
	return indexGeneric(ctx, reposDir, targetFile, name, revision, func(repoCopyDir string) error {
		if err := run.Bash(ctx, "go", "mod", "tidy").Dir(repoCopyDir).Run().Wait(); err != nil {
			return err
		}
		if err := run.Bash(ctx, "go", "mod", "vendor").Dir(repoCopyDir).Run().Wait(); err != nil {
			return err
		}
		// --repository-root=. is necessary here as the temp dir might be within a strange
		// nest of symlinks on MacOS, which confuses the repository root detection in lsif-go.
		if err := run.Bash(ctx, "lsif-go", "--repository-root=.", "-o", targetFile).Dir(repoCopyDir).Run().Wait(); err != nil {
			return err
		}

		return nil
	})
}

func indexGoWithSCIP(ctx context.Context, reposDir, targetFile, name, revision string) error {
	return indexGeneric(ctx, reposDir, targetFile, name, revision, func(repoCopyDir string) error {
		// --repository-root=. is necessary here as the temp dir might be within a strange
		// nest of symlinks on MacOS, which confuses the repository root detection in scip-go.
		if err := run.Bash(ctx, "scip-go", "--repository-root=.", "-o", targetFile).Dir(repoCopyDir).Run().Wait(); err != nil {
			return err
		}

		return nil
	})
}

func indexTypeScriptWithSCIP(ctx context.Context, reposDir, targetFile, name, revision string) error {
	return indexGeneric(ctx, reposDir, targetFile, name, revision, func(repoCopyDir string) error {
		if err := run.Bash(ctx, "yarn").Dir(repoCopyDir).Run().Wait(); err != nil {
			return err
		}
		if err := run.Bash(ctx, "scip-typescript", "index", "--output", targetFile).Dir(repoCopyDir).Run().Wait(); err != nil {
			return err
		}

		return nil
	})
}

func indexGeneric(ctx context.Context, reposDir, targetFile, name, revision string, index func(repoCopyDir string) error) error {
	if ok, err := internal.FileExists(targetFile); err != nil {
		return err
	} else if ok {
		fmt.Printf("Index for %s@%s already exists\n", name, revision)
		return nil
	}
	fmt.Printf("Indexing %s@%s\n", name, revision)

	tempDir, err := os.MkdirTemp("", "codeintel-qa")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	repoDir := filepath.Join(reposDir, name)
	repoCopyDir := filepath.Join(tempDir, name)

	if err := run.Bash(ctx, "cp", "-r", repoDir, tempDir).Run().Wait(); err != nil {
		return err
	}
	if err := run.Bash(ctx, "git", "checkout", revision).Dir(repoCopyDir).Run().Wait(); err != nil {
		return err
	}

	if err := index(repoCopyDir); err != nil {
		return err
	}

	fmt.Printf("Finished indexing %s@%s\n", name, revision)
	return nil
}
