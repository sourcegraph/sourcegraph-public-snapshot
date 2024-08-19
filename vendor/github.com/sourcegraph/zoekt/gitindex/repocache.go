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

package gitindex

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	git "github.com/go-git/go-git/v5"
)

// RepoCache is a set of repositories on the file system, named and
// stored by URL.
type RepoCache struct {
	baseDir string

	reposMu sync.Mutex
	repos   map[string]*git.Repository
}

// NewRepoCache creates a new RepoCache rooted at the given directory.
func NewRepoCache(dir string) *RepoCache {
	return &RepoCache{
		baseDir: dir,
		repos:   make(map[string]*git.Repository),
	}
}

func repoKeyStr(key string) string {
	if !strings.HasSuffix(key, ".git") {
		key += ".git"
	}
	return key
}

func repoKey(u *url.URL) string {
	return repoKeyStr(filepath.Join(u.Host, u.Path))
}

// Path returns the absolute path of the bare repository.
func Path(baseDir string, name string) string {
	key := repoKeyStr(name)
	return filepath.Join(baseDir, key)
}

func (rc *RepoCache) Path(u *url.URL) string {
	key := repoKey(u)
	return filepath.Join(rc.baseDir, key)
}

// Open opens a git repository. The cache retains a pointer to the
// repository.
func (rc *RepoCache) Open(u *url.URL) (*git.Repository, error) {
	dir := rc.Path(u)
	rc.reposMu.Lock()
	defer rc.reposMu.Unlock()

	key := repoKey(u)
	r := rc.repos[key]
	if r != nil {
		return r, nil
	}

	repo, err := git.PlainOpen(dir)
	if err == nil {
		rc.repos[key] = repo
	}
	return repo, err
}

// ListRepos returns paths to repos on disk that start with the given
// URL prefix. The paths are relative to baseDir, and typically
// include a ".git" suffix.
func ListRepos(baseDir string, u *url.URL) ([]string, error) {
	key := filepath.Join(u.Host, u.Path)

	var paths []string
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".git") && !strings.HasSuffix(path, "/.git") {
			_, err := git.PlainOpen(path)
			if err == nil {
				p, err := filepath.Rel(baseDir, path)
				if err == nil {
					paths = append(paths, p)
				}
			}
			return filepath.SkipDir
		}
		return nil
	}

	if err := filepath.Walk(filepath.Join(baseDir, key), walk); err != nil {
		return nil, err
	}
	return paths, nil
}
