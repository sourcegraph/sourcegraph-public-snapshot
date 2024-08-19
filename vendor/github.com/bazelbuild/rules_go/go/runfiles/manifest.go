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
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ManifestFile specifies the location of the runfile manifest file.  You can
// pass this as an option to New.  If unset or empty, use the value of the
// environmental variable RUNFILES_MANIFEST_FILE.
type ManifestFile string

func (f ManifestFile) new(sourceRepo SourceRepo) (*Runfiles, error) {
	m, err := f.parse()
	if err != nil {
		return nil, err
	}
	env := []string{
		manifestFileVar + "=" + string(f),
	}
	// Certain tools (e.g., Java tools) may need the runfiles directory, so try to find it even if
	// running with a manifest file.
	if strings.HasSuffix(string(f), ".runfiles_manifest") ||
		strings.HasSuffix(string(f), "/MANIFEST") ||
		strings.HasSuffix(string(f), "\\MANIFEST") {
		// Cut off either "_manifest" or "/MANIFEST" or "\\MANIFEST", all of length 9, from the end
		// of the path to obtain the runfiles directory.
		d := string(f)[:len(string(f))-len("_manifest")]
		env = append(env,
			directoryVar+"="+d,
			legacyDirectoryVar+"="+d)
	}
	r := &Runfiles{
		impl:       m,
		env:        env,
		sourceRepo: string(sourceRepo),
	}
	err = r.loadRepoMapping()
	return r, err
}

type manifest map[string]string

func (f ManifestFile) parse() (manifest, error) {
	r, err := os.Open(string(f))
	if err != nil {
		return nil, fmt.Errorf("runfiles: can’t open manifest file: %w", err)
	}
	defer r.Close()

	s := bufio.NewScanner(r)
	m := make(manifest)
	for s.Scan() {
		fields := strings.SplitN(s.Text(), " ", 2)
		if len(fields) != 2 || fields[0] == "" {
			return nil, fmt.Errorf("runfiles: bad manifest line %q in file %s", s.Text(), f)
		}
		m[fields[0]] = filepath.FromSlash(fields[1])
	}

	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("runfiles: error parsing manifest file %s: %w", f, err)
	}

	return m, nil
}

func (m manifest) path(s string) (string, error) {
	r, ok := m[s]
	if ok && r == "" {
		return "", ErrEmpty
	}
	if ok {
		return r, nil
	}

	// If path references a runfile that lies under a directory that itself is a
	// runfile, then only the directory is listed in the manifest. Look up all
	// prefixes of path in the manifest.
	for prefix := s; prefix != ""; prefix, _ = path.Split(prefix) {
		prefix = strings.TrimSuffix(prefix, "/")
		if prefixMatch, ok := m[prefix]; ok {
			return prefixMatch + filepath.FromSlash(strings.TrimPrefix(s, prefix)), nil
		}
	}

	return "", os.ErrNotExist
}
