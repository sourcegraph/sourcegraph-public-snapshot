// Copyright 2014 The Gogs Authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package GoGits - Git is a pure Go implementation of Git manipulation.
package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// idx-file
type idxFile struct {
	indexpath    string
	packpath     string
	packversion  uint32
	offsetValues map[sha1]uint64
}

// IsNotFound returns whether the error is about failing to find an object (RefNotFound, ObjectNotFound, etc).
func IsNotFound(err error) bool {
	switch err.(type) {
	case RefNotFound:
		return true
	case ObjectNotFound:
		return true
	}
	return false
}

// A Repository is the base of all other actions. If you need to lookup a
// commit, tree or blob, you do it from here.
type Repository struct {
	Path       string
	indexfiles map[string]*idxFile

	commitCache map[sha1]*Commit
	tagCache    map[sha1]*Tag
}

// Open the repository at the given path.
func OpenRepository(path string) (*Repository, error) {
	repo := new(Repository)
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	repo.Path = path
	fm, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fm.IsDir() {
		return nil, errors.New(fmt.Sprintf("%q is not a directory.", fm.Name()))
	}

	indexfiles, err := filepath.Glob(filepath.Join(path, "objects/pack/*idx"))
	if err != nil {
		return nil, err
	}
	repo.indexfiles = make(map[string]*idxFile, len(indexfiles))
	for _, indexfile := range indexfiles {
		idx, err := readIdxFile(indexfile)
		if err != nil {
			return nil, err
		}
		repo.indexfiles[indexfile] = idx
	}

	return repo, nil
}
