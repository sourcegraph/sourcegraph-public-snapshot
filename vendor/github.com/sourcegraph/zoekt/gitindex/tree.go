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
	"fmt"
	"io"
	"log"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"

	git "github.com/go-git/go-git/v5"
)

// repoWalker walks a tree, recursing into submodules.
type repoWalker struct {
	repo *git.Repository

	repoURL *url.URL
	tree    map[fileKey]BlobLocation

	// Path => SubmoduleEntry
	submodules map[string]*SubmoduleEntry

	// Path => commit SHA1
	subRepoVersions map[string]plumbing.Hash
	repoCache       *RepoCache
}

// subURL returns the URL for a submodule.
func (w *repoWalker) subURL(relURL string) (*url.URL, error) {
	if w.repoURL == nil {
		return nil, fmt.Errorf("no URL for base repo")
	}
	if strings.HasPrefix(relURL, "../") {
		u := *w.repoURL
		u.Path = path.Join(u.Path, relURL)
		return &u, nil
	}

	return url.Parse(relURL)
}

// newRepoWalker creates a new repoWalker.
func newRepoWalker(r *git.Repository, repoURL string, repoCache *RepoCache) *repoWalker {
	u, _ := url.Parse(repoURL)
	return &repoWalker{
		repo:            r,
		repoURL:         u,
		tree:            map[fileKey]BlobLocation{},
		repoCache:       repoCache,
		subRepoVersions: map[string]plumbing.Hash{},
	}
}

// parseModuleMap initializes rw.submodules.
func (rw *repoWalker) parseModuleMap(t *object.Tree) error {
	if rw.repoCache == nil {
		return nil
	}
	modEntry, _ := t.File(".gitmodules")
	if modEntry != nil {
		c, err := blobContents(&modEntry.Blob)
		if err != nil {
			return fmt.Errorf("blobContents: %w", err)
		}
		mods, err := ParseGitModules(c)
		if err != nil {
			return fmt.Errorf("ParseGitModules: %w", err)
		}
		rw.submodules = map[string]*SubmoduleEntry{}
		for _, entry := range mods {
			rw.submodules[entry.Path] = entry
		}
	}
	return nil
}

// TreeToFiles fetches the blob SHA1s for a tree. If repoCache is
// non-nil, recurse into submodules. In addition, it returns a mapping
// that indicates in which repo each SHA1 can be found.
func TreeToFiles(r *git.Repository, t *object.Tree,
	repoURL string, repoCache *RepoCache,
) (map[fileKey]BlobLocation, map[string]plumbing.Hash, error) {
	rw := newRepoWalker(r, repoURL, repoCache)

	if err := rw.parseModuleMap(t); err != nil {
		return nil, nil, fmt.Errorf("parseModuleMap: %w", err)
	}

	tw := object.NewTreeWalker(t, true, make(map[plumbing.Hash]bool))
	defer tw.Close()
	for {
		name, entry, err := tw.Next()
		if err == io.EOF {
			break
		}
		if err := rw.handleEntry(name, &entry); err != nil {
			return nil, nil, fmt.Errorf("handleEntry: %w", err)
		}
	}
	return rw.tree, rw.subRepoVersions, nil
}

func (r *repoWalker) tryHandleSubmodule(p string, id *plumbing.Hash) error {
	if err := r.handleSubmodule(p, id); err != nil {
		log.Printf("submodule %s: ignoring error %v", p, err)
	}
	return nil
}

func (r *repoWalker) handleSubmodule(p string, id *plumbing.Hash) error {
	submod := r.submodules[p]
	if submod == nil {
		return fmt.Errorf("no entry for submodule path %q", r.repoURL)
	}

	subURL, err := r.subURL(submod.URL)
	if err != nil {
		return err
	}

	subRepo, err := r.repoCache.Open(subURL)
	if err != nil {
		return err
	}

	obj, err := subRepo.CommitObject(*id)
	if err != nil {
		return err
	}
	tree, err := subRepo.TreeObject(obj.TreeHash)
	if err != nil {
		return err
	}

	r.subRepoVersions[p] = *id

	subTree, subVersions, err := TreeToFiles(subRepo, tree, subURL.String(), r.repoCache)
	if err != nil {
		return err
	}
	for k, repo := range subTree {
		r.tree[fileKey{
			SubRepoPath: filepath.Join(p, k.SubRepoPath),
			Path:        k.Path,
			ID:          k.ID,
		}] = repo
	}
	for k, v := range subVersions {
		r.subRepoVersions[filepath.Join(p, k)] = v
	}
	return nil
}

func (r *repoWalker) handleEntry(p string, e *object.TreeEntry) error {
	if e.Mode == filemode.Submodule && r.repoCache != nil {
		if err := r.tryHandleSubmodule(p, &e.Hash); err != nil {
			return fmt.Errorf("submodule %s: %v", p, err)
		}
	}

	switch e.Mode {
	case filemode.Regular, filemode.Executable, filemode.Symlink:
	default:
		return nil
	}

	r.tree[fileKey{
		Path: p,
		ID:   e.Hash,
	}] = BlobLocation{
		Repo: r.repo,
		URL:  r.repoURL,
	}
	return nil
}

// fileKey describes a blob at a location in the final tree. We also
// record the subrepository from where it came.
type fileKey struct {
	SubRepoPath string
	Path        string
	ID          plumbing.Hash
}

func (k *fileKey) FullPath() string {
	return filepath.Join(k.SubRepoPath, k.Path)
}

// BlobLocation holds data where a blob can be found.
type BlobLocation struct {
	Repo *git.Repository
	URL  *url.URL
}

func (l *BlobLocation) Blob(id *plumbing.Hash) ([]byte, error) {
	blob, err := l.Repo.BlobObject(*id)
	if err != nil {
		return nil, err
	}
	return blobContents(blob)
}
