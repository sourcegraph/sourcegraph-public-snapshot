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

// Package runfiles provides access to Bazel runfiles.
//
// # Usage
//
// This package has two main entry points, the global functions Rlocation and Env,
// and the Runfiles type.
//
// # Global functions
//
// For simple use cases that don’t require hermetic behavior, use the Rlocation and
// Env functions to access runfiles.  Use Rlocation to find the filesystem location
// of a runfile, and use Env to obtain environmental variables to pass on to
// subprocesses.
//
// # Runfiles type
//
// If you need hermetic behavior or want to change the runfiles discovery
// process, use New to create a Runfiles object.  New accepts a few options to
// change the discovery process.  Runfiles objects have methods Rlocation and Env,
// which correspond to the package-level functions.  On Go 1.16, *Runfiles
// implements fs.FS, fs.StatFS, and fs.ReadFileFS.
package runfiles

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	directoryVar       = "RUNFILES_DIR"
	legacyDirectoryVar = "JAVA_RUNFILES"
	manifestFileVar    = "RUNFILES_MANIFEST_FILE"
)

type repoMappingKey struct {
	sourceRepo             string
	targetRepoApparentName string
}

// Runfiles allows access to Bazel runfiles.  Use New to create Runfiles
// objects; the zero Runfiles object always returns errors.  See
// https://docs.bazel.build/skylark/rules.html#runfiles for some information on
// Bazel runfiles.
type Runfiles struct {
	// We don’t need concurrency control since Runfiles objects are
	// immutable once created.
	impl        runfiles
	env         []string
	repoMapping map[repoMappingKey]string
	sourceRepo  string
}

const noSourceRepoSentinel = "_not_a_valid_repository_name"

// New creates a given Runfiles object.  By default, it uses os.Args and the
// RUNFILES_MANIFEST_FILE and RUNFILES_DIR environmental variables to find the
// runfiles location.  This can be overwritten by passing some options.
//
// See section “Runfiles discovery” in
// https://docs.google.com/document/d/e/2PACX-1vSDIrFnFvEYhKsCMdGdD40wZRBX3m3aZ5HhVj4CtHPmiXKDCxioTUbYsDydjKtFDAzER5eg7OjJWs3V/pub.
func New(opts ...Option) (*Runfiles, error) {
	var o options
	o.sourceRepo = noSourceRepoSentinel
	for _, a := range opts {
		a.apply(&o)
	}

	if o.sourceRepo == noSourceRepoSentinel {
		o.sourceRepo = SourceRepo(CallerRepository())
	}

	if o.manifest == "" {
		o.manifest = ManifestFile(os.Getenv(manifestFileVar))
	}
	if o.manifest != "" {
		return o.manifest.new(o.sourceRepo)
	}

	if o.directory == "" {
		o.directory = Directory(os.Getenv(directoryVar))
	}
	if o.directory != "" {
		return o.directory.new(o.sourceRepo)
	}

	if o.program == "" {
		o.program = ProgramName(os.Args[0])
	}
	manifest := ManifestFile(o.program + ".runfiles_manifest")
	if stat, err := os.Stat(string(manifest)); err == nil && stat.Mode().IsRegular() {
		return manifest.new(o.sourceRepo)
	}

	dir := Directory(o.program + ".runfiles")
	if stat, err := os.Stat(string(dir)); err == nil && stat.IsDir() {
		return dir.new(o.sourceRepo)
	}

	return nil, errors.New("runfiles: no runfiles found")
}

// Rlocation returns the (relative or absolute) path name of a runfile.
// The runfile name must be a runfile-root relative path, using the slash (not
// backslash) as directory separator. It is typically of the form
// "repo/path/to/pkg/file".
//
// If r is the zero Runfiles object, Rlocation always returns an error. If the
// runfiles manifest maps s to an empty name (indicating an empty runfile not
// present in the filesystem), Rlocation returns an error that wraps ErrEmpty.
//
// See section “Library interface” in
// https://docs.google.com/document/d/e/2PACX-1vSDIrFnFvEYhKsCMdGdD40wZRBX3m3aZ5HhVj4CtHPmiXKDCxioTUbYsDydjKtFDAzER5eg7OjJWs3V/pub.
func (r *Runfiles) Rlocation(path string) (string, error) {
	if r.impl == nil {
		return "", errors.New("runfiles: uninitialized Runfiles object")
	}

	if path == "" {
		return "", errors.New("runfiles: path may not be empty")
	}
	if err := isNormalizedPath(path); err != nil {
		return "", err
	}

	// See https://github.com/bazelbuild/bazel/commit/b961b0ad6cc2578b98d0a307581e23e73392ad02
	if strings.HasPrefix(path, `\`) {
		return "", fmt.Errorf("runfiles: path %q is absolute without a drive letter", path)
	}
	if filepath.IsAbs(path) {
		return path, nil
	}

	mappedPath := path
	split := strings.SplitN(path, "/", 2)
	if len(split) == 2 {
		key := repoMappingKey{r.sourceRepo, split[0]}
		if targetRepoDirectory, exists := r.repoMapping[key]; exists {
			mappedPath = targetRepoDirectory + "/" + split[1]
		}
	}

	p, err := r.impl.path(mappedPath)
	if err != nil {
		return "", Error{path, err}
	}
	return p, nil
}

func isNormalizedPath(s string) error {
	if strings.HasPrefix(s, "../") || strings.Contains(s, "/../") || strings.HasSuffix(s, "/..") {
		return fmt.Errorf(`runfiles: path %q must not contain ".." segments`, s)
	}
	if strings.HasPrefix(s, "./") || strings.Contains(s, "/./") || strings.HasSuffix(s, "/.") {
		return fmt.Errorf(`runfiles: path %q must not contain "." segments`, s)
	}
	if strings.Contains(s, "//") {
		return fmt.Errorf(`runfiles: path %q must not contain "//"`, s)
	}
	return nil
}

// loadRepoMapping loads the repo mapping (if it exists) using the impl.
// This mutates the Runfiles object, but is idempotent.
func (r *Runfiles) loadRepoMapping() error {
	repoMappingPath, err := r.impl.path(repoMappingRlocation)
	// If Bzlmod is disabled, the repository mapping manifest isn't created, so
	// it is not an error if it is missing.
	if err != nil {
		return nil
	}
	r.repoMapping, err = parseRepoMapping(repoMappingPath)
	// If the repository mapping manifest exists, it must be valid.
	return err
}

// Env returns additional environmental variables to pass to subprocesses.
// Each element is of the form “key=value”.  Pass these variables to
// Bazel-built binaries so they can find their runfiles as well.  See the
// Runfiles example for an illustration of this.
//
// The return value is a newly-allocated slice; you can modify it at will.  If
// r is the zero Runfiles object, the return value is nil.
func (r *Runfiles) Env() []string {
	return r.env
}

// WithSourceRepo returns a Runfiles instance identical to the current one,
// except that it uses the given repository's repository mapping when resolving
// runfiles paths.
func (r *Runfiles) WithSourceRepo(sourceRepo string) *Runfiles {
	if r.sourceRepo == sourceRepo {
		return r
	}
	clone := *r
	clone.sourceRepo = sourceRepo
	return &clone
}

// Option is an option for the New function to override runfiles discovery.
type Option interface {
	apply(*options)
}

// ProgramName is an Option that sets the program name. If not set, New uses
// os.Args[0].
type ProgramName string

// SourceRepo is an Option that sets the canonical name of the repository whose
// repository mapping should be used to resolve runfiles paths. If not set, New
// uses the repository containing the source file from which New is called.
// Use CurrentRepository to get the name of the current repository.
type SourceRepo string

// Error represents a failure to look up a runfile.
type Error struct {
	// Runfile name that caused the failure.
	Name string

	// Underlying error.
	Err error
}

// Error implements error.Error.
func (e Error) Error() string {
	return fmt.Sprintf("runfile %s: %s", e.Name, e.Err.Error())
}

// Unwrap returns the underlying error, for errors.Unwrap.
func (e Error) Unwrap() error { return e.Err }

// ErrEmpty indicates that a runfile isn’t present in the filesystem, but
// should be created as an empty file if necessary.
var ErrEmpty = errors.New("empty runfile")

type options struct {
	program    ProgramName
	manifest   ManifestFile
	directory  Directory
	sourceRepo SourceRepo
}

func (p ProgramName) apply(o *options)  { o.program = p }
func (m ManifestFile) apply(o *options) { o.manifest = m }
func (d Directory) apply(o *options)    { o.directory = d }
func (sr SourceRepo) apply(o *options)  { o.sourceRepo = sr }

type runfiles interface {
	path(string) (string, error)
}

// The runfiles root symlink under which the repository mapping can be found.
// https://cs.opensource.google/bazel/bazel/+/1b073ac0a719a09c9b2d1a52680517ab22dc971e:src/main/java/com/google/devtools/build/lib/analysis/Runfiles.java;l=424
const repoMappingRlocation = "_repo_mapping"

// Parses a repository mapping manifest file emitted with Bzlmod enabled.
func parseRepoMapping(path string) (map[repoMappingKey]string, error) {
	r, err := os.Open(path)
	if err != nil {
		// The repo mapping manifest only exists with Bzlmod, so it's not an
		// error if it's missing. Since any repository name not contained in the
		// mapping is assumed to be already canonical, an empty map is
		// equivalent to not applying any mapping.
		return nil, nil
	}
	defer r.Close()

	// Each line of the repository mapping manifest has the form:
	// canonical name of source repo,apparent name of target repo,target repo runfiles directory
	// https://cs.opensource.google/bazel/bazel/+/1b073ac0a719a09c9b2d1a52680517ab22dc971e:src/main/java/com/google/devtools/build/lib/analysis/RepoMappingManifestAction.java;l=117
	s := bufio.NewScanner(r)
	repoMapping := make(map[repoMappingKey]string)
	for s.Scan() {
		fields := strings.SplitN(s.Text(), ",", 3)
		if len(fields) != 3 {
			return nil, fmt.Errorf("runfiles: bad repo mapping line %q in file %s", s.Text(), path)
		}
		repoMapping[repoMappingKey{fields[0], fields[1]}] = fields[2]
	}

	if err = s.Err(); err != nil {
		return nil, fmt.Errorf("runfiles: error parsing repo mapping file %s: %w", path, err)
	}

	return repoMapping, nil
}
