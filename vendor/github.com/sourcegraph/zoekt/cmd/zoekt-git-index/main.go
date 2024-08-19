// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/sourcegraph/zoekt/cmd"
	"github.com/sourcegraph/zoekt/ctags"
	"github.com/sourcegraph/zoekt/gitindex"
)

func run() int {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to `file`")

	allowMissing := flag.Bool("allow_missing_branches", false, "allow missing branches.")
	submodules := flag.Bool("submodules", true, "if set to false, do not recurse into submodules")
	branchesStr := flag.String("branches", "HEAD", "git branches to index.")
	branchPrefix := flag.String("prefix", "refs/heads/", "prefix for branch names")

	incremental := flag.Bool("incremental", true, "only index changed repositories")
	repoCacheDir := flag.String("repo_cache", "", "directory holding bare git repos, named by URL. "+
		"this is used to find repositories for submodules. "+
		"It also affects name if the indexed repository is under this directory.")
	isDelta := flag.Bool("delta", false, "whether we should use delta build")
	deltaShardNumberFallbackThreshold := flag.Uint64("delta_threshold", 0, "upper limit on the number of preexisting shards that can exist before attempting a delta build (0 to disable fallback behavior)")
	offlineRanking := flag.String("offline_ranking", "", "the name of the file that contains the ranking info.")
	offlineRankingVersion := flag.String("offline_ranking_version", "", "a version string identifying the contents in offline_ranking.")
	languageMap := flag.String("language_map", "", "a mapping between a language and its ctags processor (a:0,b:3).")
	flag.Parse()

	// Tune GOMAXPROCS to match Linux container CPU quota.
	_, _ = maxprocs.Set()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if *repoCacheDir != "" {
		dir, err := filepath.Abs(*repoCacheDir)
		if err != nil {
			log.Fatalf("Abs: %v", err)
		}
		*repoCacheDir = dir
	}
	opts := cmd.OptionsFromFlags()
	opts.IsDelta = *isDelta
	opts.DocumentRanksPath = *offlineRanking
	opts.DocumentRanksVersion = *offlineRankingVersion

	var branches []string
	if *branchesStr != "" {
		branches = strings.Split(*branchesStr, ",")
	}

	gitRepos := map[string]string{}
	for _, repoDir := range flag.Args() {
		repoDir, err := filepath.Abs(repoDir)
		if err != nil {
			log.Fatal(err)
		}
		repoDir = filepath.Clean(repoDir)

		name := strings.TrimSuffix(repoDir, "/.git")
		if *repoCacheDir != "" && strings.HasPrefix(name, *repoCacheDir) {
			name = strings.TrimPrefix(name, *repoCacheDir+"/")
			name = strings.TrimSuffix(name, ".git")
		} else {
			name = strings.TrimSuffix(filepath.Base(name), ".git")
		}
		gitRepos[repoDir] = name
	}

	opts.LanguageMap = make(ctags.LanguageMap)
	for _, mapping := range strings.Split(*languageMap, ",") {
		m := strings.Split(mapping, ":")
		if len(m) != 2 {
			continue
		}
		opts.LanguageMap[m[0]] = ctags.StringToParser(m[1])
	}

	exitStatus := 0
	for dir, name := range gitRepos {
		opts.RepositoryDescription.Name = name
		gitOpts := gitindex.Options{
			BranchPrefix:                      *branchPrefix,
			Incremental:                       *incremental,
			Submodules:                        *submodules,
			RepoCacheDir:                      *repoCacheDir,
			AllowMissingBranch:                *allowMissing,
			BuildOptions:                      *opts,
			Branches:                          branches,
			RepoDir:                           dir,
			DeltaShardNumberFallbackThreshold: *deltaShardNumberFallbackThreshold,
		}

		if _, err := gitindex.IndexGitRepo(gitOpts); err != nil {
			log.Printf("indexGitRepo(%s, delta=%t): %v", dir, gitOpts.BuildOptions.IsDelta, err)
			exitStatus = 1
		}
	}

	return exitStatus
}

func main() {
	exitStatus := run()
	os.Exit(exitStatus)
}
