// Copyright 2020, 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runfiles

import (
	"regexp"
	"runtime"
	"sync"
)

// Rlocation returns the absolute path name of a runfile.  The runfile name must be
// a relative path, using the slash (not backslash) as directory separator.  If
// the runfiles manifest maps s to an empty name (indicating an empty runfile
// not present in the filesystem), Rlocation returns an error that wraps ErrEmpty.
func Rlocation(path string) (string, error) {
	return RlocationFrom(path, CallerRepository())
}

func RlocationFrom(path string, sourceRepo string) (string, error) {
	r, err := g.get()
	if err != nil {
		return "", err
	}
	return r.WithSourceRepo(sourceRepo).Rlocation(path)
}

// Env returns additional environmental variables to pass to subprocesses.
// Each element is of the form “key=value”.  Pass these variables to
// Bazel-built binaries so they can find their runfiles as well.  See the
// Runfiles example for an illustration of this.
//
// The return value is a newly-allocated slice; you can modify it at will.
func Env() ([]string, error) {
	r, err := g.get()
	if err != nil {
		return nil, err
	}
	return r.Env(), nil
}

var legacyExternalGeneratedFile = regexp.MustCompile(`^bazel-out[/][^/]+/bin/external/([^/]+)/`)
var legacyExternalFile = regexp.MustCompile(`^external/([^/]+)/`)

// CurrentRepository returns the canonical name of the Bazel repository that
// contains the source file of the caller of CurrentRepository.
func CurrentRepository() string {
	return callerRepository(1)
}

// CallerRepository returns the canonical name of the Bazel repository that
// contains the source file of the caller of the function that itself calls
// CallerRepository.
func CallerRepository() string {
	return callerRepository(2)
}

func callerRepository(skip int) string {
	_, file, _, _ := runtime.Caller(skip + 1)
	if match := legacyExternalGeneratedFile.FindStringSubmatch(file); match != nil {
		return match[1]
	}
	if match := legacyExternalFile.FindStringSubmatch(file); match != nil {
		return match[1]
	}
	// If a file is not in an external repository, it is in the main repository,
	// which has the empty string as its canonical name.
	return ""
}

type global struct {
	once     sync.Once
	runfiles *Runfiles
	err      error
}

func (g *global) get() (*Runfiles, error) {
	g.once.Do(g.init)
	return g.runfiles, g.err
}

func (g *global) init() {
	g.runfiles, g.err = New()
}

var g global
